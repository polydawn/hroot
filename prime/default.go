package prime

//Default options to pass to docker

//Representing JSON in Golang is a little convoluted:
//	map[string]interface{} is object
//	[]string is an array of strings, etc

// `socat` is a completely excellent tool for seeing what docker in the wild does on the wire:
//    socat -v UNIX-LISTEN:listener,fork UNIX:/var/run/docker.sock
// it's much like wireshark, but here it's forwarding a unix socket and logs just exactly what you wish it did flat on stderr.

//Default options for creating a docker
var DefaultCreate = map[string]interface{}{
	"Hostname": "",
	"User": "",
	"Memory": 0,
	"MemorySwap": 0,
	"CpuShares": 0,
	"AttachStdin": false,
	"AttachStdout": false,
	"AttachStderr": false,
	"PortSpecs": []interface{}{},
	"Tty": false,
	"OpenStdin": false,
	"StdinOnce": false,
	"Env": nil,
	"Cmd": []interface{}{"/service/launch.sh"},
	"Dns": nil,
	"Image": nil, // this absolutely must be overriden
	"Volumes": map[string]interface{}{},
	"VolumesFrom": "",
	"Entrypoint": []string{},
	"NetworkDisabled": false,
	"Privileged": false,
	"WorkingDir": "/",
}

//Default options for running a docker
var DefaultStart = map[string]interface{}{
	"Binds": []interface{}{},
	"ContainerIDFile": "",
}

//Default options for provisioning and launching a service
var DefaultService = map[string]interface{}{
	"Launch": nil,       //Normal operation
	"Build": nil,        //Provision
	"Persistent": false, //Delete container after running?
}

//Everything together
var DefaultConfig = map[string]interface{}{
	"Create": DefaultCreate,
	"Start":  DefaultStart,
	"Service": DefaultService,
}
