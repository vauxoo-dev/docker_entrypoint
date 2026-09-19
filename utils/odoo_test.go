package utils

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/ini.v1"
)

func TestFilterStrings(t *testing.T) {
	executed := false
	_ = FilterStrings([]string{}, func([]string) map[string]string {
		executed = true
		return nil
	})
	if !assert.True(t, executed) {
		t.Errorf("The function FilterStrings was not called.")
	}
}

func TestGetMainRepoPath(t *testing.T) {
	values := []struct {
		mainRepo string
		expected string
	}{
		{"extra_addons/customer", "/home/odoo/instance/extra_addons/customer"},
		{"", "/home/odoo/instance/odoo"},
	}
	for _, v := range values {
		t.Setenv("MAIN_REPO_PATH", v.mainRepo)
		assert.Equal(t, v.expected, GetMainRepoPath(),
			"MAIN_REPO_PATH must be resolved under the instance directory.")
	}
}

func TestGetMainRepoBranch(t *testing.T) {
	repoPath := t.TempDir()
	if err := exec.Command("git", "init", "-b", "18.0-dev1", repoPath).Run(); err != nil {
		t.Skipf("git is not available: %s", err)
	}
	assert.Equal(t, "18.0-dev1", GetMainRepoBranch(repoPath),
		"The branch of the repository must be reported.")
	assert.Equal(t, "", GetMainRepoBranch(t.TempDir()),
		"A directory without a repository must report no branch.")
}

func TestGetImageTag(t *testing.T) {
	t.Setenv("MAIN_APP", "customer")
	t.Setenv("VERSION", "19.0")
	repoPath := t.TempDir()
	if err := exec.Command("git", "init", "-b", "19.0", repoPath).Run(); err != nil {
		t.Skipf("git is not available: %s", err)
	}
	if err := exec.Command("git", "-C", repoPath, "-c", "user.email=a@b", "-c", "user.name=a",
		"commit", "--allow-empty", "-m", "empty").Run(); err != nil {
		t.Fatal(err)
	}
	commit := RunGit(repoPath, "rev-parse", "--short", "HEAD")
	assert.Equal(t, "customer-19.0-"+commit, GetImageTag(repoPath),
		"The tag must be built from MAIN_APP, VERSION and the commit.")
	assert.Equal(t, "", GetImageTag(t.TempDir()),
		"Without a repository there is no tag to report.")
	t.Setenv("MAIN_APP", "")
	assert.Equal(t, "", GetImageTag(repoPath),
		"Without MAIN_APP there is no tag to report.")
}

func TestGetSentryEnvironment(t *testing.T) {
	// The main repository does not exist while the tests run, so the branch lookup finds nothing
	// and the environment falls back to VERSION.
	values := []struct {
		instanceType string
		version      string
		expected     string
	}{
		{"production", "19.0", "production-19.0"},
		{"production", "18.0", "production-18.0"},
		{"develop", "", "develop"},
	}
	for _, v := range values {
		t.Setenv("VERSION", v.version)
		assert.Equal(t, v.expected, GetSentryEnvironment(v.instanceType),
			"The environment must join the stage and the version.")
	}
}

func TestSplitEnvVars(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"var=111", "odoorc_var=123", "CAPS_VAR=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"var": "111", "odoorc_var": "123", "CAPS_VAR": "wer", "ODOORC_INCAPS": "caps"},
		},
	}
	for _, v := range values {
		res := SplitEnvVars(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestDefaultConverter(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"var=111", "odoorc_var=123", "CAPS_VAR=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"var": "111", "odoorc_var": "123", "CAPS_VAR": "wer", "ODOORC_INCAPS": "caps"},
		},
	}
	for _, v := range values {
		res := DefaultConverter(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestOdoorcConverter(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"var=111", "odoorc_var=123", "CAPS_VAR=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"var": "123", "incaps": "caps"},
		},
	}
	for _, v := range values {
		res := OdoorcConverter(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestGetOdooUser(t *testing.T) {
	res := GetOdooUser()
	assert.Equal(t, "odoo", res)
}

func TestGetConfigFile(t *testing.T) {
	res := GetConfigFile()
	assert.Equal(t, "/home/odoo/.openerp_serverrc", res)
	err := os.Setenv("ODOO_CONFIG_FILE", "/etc/odoo.conf")
	assert.NoError(t, err)
	res = GetConfigFile()
	assert.Equal(t, "/etc/odoo.conf", res)
	os.Unsetenv("ODOO_CONFIG_FILE")
}

func TestGetInstanceType(t *testing.T) {
	_, err := GetInstanceType()
	assert.Errorf(t, err, "cannot determine the instance type, env vars INSTANCE_TYPE and/or ODOO_STAGE 'must' be defined and match")

	err = os.Setenv("INSTANCE_TYPE", "test")
	assert.NoError(t, err)
	res, err := GetInstanceType()
	assert.NoError(t, err)
	assert.Equal(t, "test", res)

	err = os.Setenv("ODOO_STAGE", "dev")
	assert.NoError(t, err)
	_, err = GetInstanceType()
	assert.Errorf(t, err, "cannot determine the instance type, env vars INSTANCE_TYPE and ODOO_STAGE 'must' match")
}

func TestParseVersion(t *testing.T) {
	test := []struct {
		given string
		want  string
	}{
		{
			given: "17.0",
			want:  "v17.0",
		},
		{
			given: "saas-8",
			want:  "v8",
		},
		{
			given: "saas-17.4",
			want:  "v17.4",
		},
	}

	for _, c := range test {
		assert.Equal(t, c.want, parseVersion(c.given))
	}
}

//func TestUpdateSentry(t *testing.T) {
//	values := []struct{
//		input map[string]string
//		instanceType string
//		expected map[string]string
//	}{
//		{
//			map[string]string{"sentry_enabled": "true"},
//			"develop",
//			map[string]string{"sentry_enabled": "true", "sentry_odoo_dir": "/home/odoo/instance/odoo", "sentry_environment": "develop"},
//		},
//		{
//			map[string]string{"sentry_enabled": "false"},
//			"test",
//			map[string]string{"sentry_enabled": "false"},
//		},
//		{
//			map[string]string{"sentry_enabled": "True"},
//			"production",
//			map[string]string{"sentry_enabled": "True", "sentry_odoo_dir": "/home/odoo/instance/odoo", "sentry_environment": "production"},
//		},
//		{
//			map[string]string{"not_sentry": "True"},
//			"production",
//			map[string]string{"not_sentry": "True"},
//		},
//	}
//	for _, v := range values {
//		UpdateSentry(v.input, v.instanceType)
//		if !assert.Equal(t, v.expected, v.input) {
//			t.Errorf("Got: %+v, expected: %+v", v.input, v.expected)
//		}
//	}
//}

// loadConfig builds an ini.File from the literal a rendered .odoorc would contain.
func loadConfig(t *testing.T, content string) *ini.File {
	t.Helper()
	config, err := ini.Load([]byte(content))
	assert.NoError(t, err)
	return config
}

// TestSentrySectionPicksTheDeclaredSection covers the whole point of the change: the keys have to
// be completed where the module reads them, which since 19.0 is a [sentry] section when the image
// declares one, and [options] otherwise.
func TestSentrySectionPicksTheDeclaredSection(t *testing.T) {
	withSection := loadConfig(t, "[options]\ndb_host = localhost\n\n[sentry]\nsentry_enabled = true\n")
	assert.Equal(t, "sentry", SentrySection(withSection).Name())

	withoutSection := loadConfig(t, "[options]\nsentry_enabled = true\n")
	assert.Equal(t, "options", SentrySection(withoutSection).Name())
}

// TestSentryValueReadsEitherSection: sentry_enabled moves to [sentry] as soon as that section
// declares it, so looking only at [options] would make UpdateSentry skip the whole configuration.
func TestSentryValueReadsEitherSection(t *testing.T) {
	config := loadConfig(t, "[options]\nsentry_dist = from-options\n\n[sentry]\nsentry_enabled = true\n")
	assert.Equal(t, "true", SentryValue(config, "sentry_enabled"))
	assert.Equal(t, "from-options", SentryValue(config, "sentry_dist"))
	assert.Equal(t, "", SentryValue(config, "sentry_release"))
}

func TestUpdateSentry(t *testing.T) {
	t.Setenv("MAIN_REPO_PATH", "extra_addons/ircodoo")
	t.Setenv("VERSION", "19.0")
	mainRepo := "/home/odoo/instance/extra_addons/ircodoo"

	t.Run("fills the declared section and leaves options alone", func(t *testing.T) {
		config := loadConfig(t, "[options]\ndb_host = localhost\n\n[sentry]\nsentry_enabled = true\nsentry_odoo_dir =\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, mainRepo, config.Section("sentry").Key("sentry_odoo_dir").Value())
		assert.Equal(t, "develop-19.0", config.Section("sentry").Key("sentry_environment").Value())
		assert.False(t, config.Section("options").HasKey("sentry_odoo_dir"),
			"nothing may be written to the section the module does not read")
	})

	t.Run("keeps writing to options when no section is declared", func(t *testing.T) {
		config := loadConfig(t, "[options]\nsentry_enabled = true\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, mainRepo, config.Section("options").Key("sentry_odoo_dir").Value())
		_, err := config.GetSection("sentry")
		assert.Error(t, err, "no section is created out of nothing")
	})

	t.Run("a given odoo_dir wins over the derived one", func(t *testing.T) {
		config := loadConfig(t, "[sentry]\nsentry_enabled = true\nsentry_odoo_dir = /somewhere/else\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, "/somewhere/else", config.Section("sentry").Key("sentry_odoo_dir").Value())
	})

	t.Run("disabled and unparseable are left untouched", func(t *testing.T) {
		for _, value := range []string{"false", "not-a-bool"} {
			config := loadConfig(t, "[sentry]\nsentry_enabled = "+value+"\n")
			UpdateSentry(config, "develop")
			assert.False(t, config.Section("sentry").HasKey("sentry_environment"),
				"nothing is derived when sentry is off, value %q", value)
		}
	})

	t.Run("no sentry_enabled anywhere is a no-op", func(t *testing.T) {
		config := loadConfig(t, "[options]\nnot_sentry = true\n")
		UpdateSentry(config, "production")
		assert.False(t, config.Section("options").HasKey("sentry_odoo_dir"))
	})
}

// TestUpdateSentryReconcilesTheDisableFlag covers the two ways the flag and the module disagree.
// The module does not parse sentry_enabled, it tests the string for truth, so "false" reads as on
// there while strconv.ParseBool reads it as off here.
func TestUpdateSentryReconcilesTheDisableFlag(t *testing.T) {
	t.Setenv("MAIN_REPO_PATH", "extra_addons/ircodoo")
	t.Setenv("VERSION", "19.0")

	t.Run("false in the section is emptied so both readers agree", func(t *testing.T) {
		config := loadConfig(t, "[sentry]\nsentry_enabled = False\nsentry_dsn = https://k@example/1\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, "", config.Section("sentry").Key("sentry_enabled").Value(),
			"a non-empty value is truthy to the module and would switch Sentry on")
		assert.False(t, config.Section("sentry").HasKey("sentry_environment"),
			"nothing is derived for a disabled instance")
	})

	t.Run("false in options is left alone", func(t *testing.T) {
		config := loadConfig(t, "[options]\nsentry_enabled = False\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, "False", config.Section("options").Key("sentry_enabled").Value(),
			"versions reading [options] do not have this problem and are not touched")
	})

	t.Run("the flag is written into the section the module reads", func(t *testing.T) {
		// A file declaring the section without the flag: UpdateFromVars appends the unmatched
		// ODOORC_SENTRY_ENABLED to [options], and the module would never see it.
		config := loadConfig(t, "[options]\nsentry_enabled = True\n\n[sentry]\nsentry_dsn = https://k@example/1\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, "True", config.Section("sentry").Key("sentry_enabled").Value())
	})

	t.Run("a given environment is kept", func(t *testing.T) {
		config := loadConfig(t, "[sentry]\nsentry_enabled = true\nsentry_environment = from-deployv\n")
		UpdateSentry(config, "develop")
		assert.Equal(t, "from-deployv", config.Section("sentry").Key("sentry_environment").Value())
	})
}
