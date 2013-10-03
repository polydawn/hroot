/*
	The crocker package provides an abstract container system.

	Crocker is based on wrapping docker, and internally, operates by exec'ing docker commands and the CLI interface.
	The Crocker API conceals the fact that it uses the exec cli, and may in the future be transparently reimplemented to use other docker APIs.
*/
package crocker
