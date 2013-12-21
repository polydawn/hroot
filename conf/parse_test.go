package conf

import (
	"path/filepath"
	"testing"
	"github.com/coocood/assrt"
)

func parser() *TomlConfigParser {
	return &TomlConfigParser{}
}

func TestTomlParser(t *testing.T) {
	//Testing library, current directory, and some config strings
	assert := assrt.NewAssert(t)
	cwd, _ := filepath.Abs(".")
	nwd, _ := filepath.Abs("..")
	f1, f2, f3, settings := "", "", "", "[settings]\n"


	//
	//	Default config
	//
	conf := parser().GetConfig()
	assert.Equal(DefaultConfiguration, *conf)


	//
	//	Basic settings
	//
	f1 = `
		# Custom DNS servers
		dns = [ "8.8.8.8", "8.8.4.4" ]

		# What container folder to start in
		folder = "/docket"

		# Interactive mode
		attach = true

		# Delete the container after running
		purge = true
	`
	conf = parser().
		AddConfig(settings + f1, ".").
		GetConfig()
	expect := DefaultConfiguration //Use default fields with a few exceptions
	expect.Settings.DNS = []string{ "8.8.8.8", "8.8.4.4" }
	expect.Settings.Folder = "/docket"
	expect.Settings.Attach = true
	expect.Settings.Purge = true
	assert.Equal(expect, *conf)


	//
	//	Mount localizing
	//
	f2 = `
		# Folder mounts
		#	(host folder, container folder, 'ro' or 'rw' permissions)
		mounts = [
			[ ".../", "/boxen",    "ro"],  # The top folder
			[ "./",   "/docket",   "rw"],  # The current folder
		]
	`
	conf = parser().
		AddConfig(settings + f1 + f2, "..").
		GetConfig()
	expect.Settings.Mounts = [][]string{
		[]string{ nwd, "/boxen",  "ro" },
		[]string{ cwd, "/docket", "rw" },
	}
	assert.Equal(expect, *conf)


	//
	//	Settings override
	//
	f3 = `
		folder = "/home"
	`
	conf = parser().
		AddConfig(settings + f1 + f2, "..").
		AddConfig(settings + f3, ".").
		GetConfig()
	expect.Settings.Folder = "/home"
	assert.Equal(expect, *conf)

}
