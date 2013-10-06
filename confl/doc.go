/*
	This package contains the config file loading system used by dockctrl.
	It produces crocker.ContainerConfig from TOML-formatted configuration files.

	Config files can arranged in nested directories, will loaded recursively, with config values accumulating 
	and the deeper config files overriding the values from the shallower files, providing a simple
	structure for inheriting common configuration.

	This package isolates crocker.ContainerConfig from any specific knowledge of TOML (admittedly, 
	crocker.ContainerConfig is annotated to help the toml loader; but it does *not* have a
	compile-time dep on a toml library).
*/
package confl
