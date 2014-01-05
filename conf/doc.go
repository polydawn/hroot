/*
	This package contains the config file loading system used by hroot.
	It produces conf.Settings from TOML-formatted configuration files.

	Config files can arranged in nested directories, will loaded recursively, with config values accumulating
	and the deeper config files overriding the values from the shallower files, providing a simple
	structure for inheriting common configuration.

	This package isolates conf.Settings from any specific knowledge of TOML (admittedly,
	conf.Settings is annotated to help the toml loader; but it does *not* have a
	compile-time dep on a toml library).
*/
package conf
