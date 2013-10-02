package prime

import (
	. "fmt"
	"io/ioutil"
	"strings"
)

const FileName = "docker.json"


//Recursively finds configuration files and loads them top-down.
//This lets you have a base configuration in the parent directory and override it for specific containers.
func loadConfig(dir string, confOverride map[string]interface{}) (map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	file, stack, config := FileName, new(Stack), DefaultConfig

	//recurse up the file tree looking for configuration
	for {
		data, err := ioutil.ReadFile(dir+"/"+file)
		if err != nil {
			break
		} else {
			stack.Push(data)
			file = "../" + file
		}
	}

	if stack.Len() > 0 {
		Println("Loaded", stack.Len(), "config files.")
	}

	//Apply the configuration file(s)
	for stack.Len() > 0 {
		content := decodeJSON(stack.Pop().([]byte))
		override(content, config)
	}

	//If there are any final overrides, load them
	override(config, confOverride)

	//Parse out volumes/binds and fix some weird docker repetition issues.
	//Also handles relative paths for you.
	confCreateVolumes := config["Create"].(map[string]interface{})["Volumes"].(map[string]interface{})
	confStartBinds := config["Start"].(map[string]interface{})["Binds"].([]interface{})
	for k, v := range confStartBinds {
		chunks := strings.SplitN(v.(string), ":", 3)
		// oh my god we seriously need a real struct to put logic on instead of shitting uncontrollably over an untyped wad of maybe-strings
		if !strings.HasPrefix(chunks[0], "/") {
			if len(chunks) == 3 {
				confStartBinds[k] = Sprintf("%s/%s:%s:%s", dir, chunks[0], chunks[1], chunks[2])
			} else {
				confStartBinds[k] = Sprintf("%s/%s:%s", dir, chunks[0], chunks[1])
			}
		}
		confCreateVolumes[chunks[1]] = struct{}{}
	}

	// Println("----")
	// debugJSON(config)

	//Cast here to keep ugly type assertions out of the main logic
	return config["Create"].(map[string]interface {}), config["Start"].(map[string]interface {}), config["Service"].(map[string]interface {})
}

//Loads a configuration object, overriding the base
func override(inc, base map[string]interface {}) {
	for k, v := range inc {
		switch vv := v.(type) {
		case nil, bool, string, int, float64:
			// Println("Loading", vv, k)
			base[k] = inc[k]
		case []interface{}:
			base[k] = append(base[k].([]interface{}), vv...)
		case map[string]interface {}:
			override(inc[k].(map[string]interface {}), base[k].(map[string]interface {})) //recurse!
		default:
			Println("Warning: ignoring unknown", vv, k)
		}
	}
}
