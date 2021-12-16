package plugins

import "fmt"

// PluginFunc returns a new instance of a Plugin
type PluginFunc func() Plugin

// Registry contains all the registered plugins
var Registry = make(map[string]PluginFunc)

// Register adds a given plugin to the registry
func Register(f PluginFunc) {
	tmp := f()
	n := tmp.Name()
	if _, ok := Registry[n]; ok {
		msg := fmt.Sprintf("Plugin named %s already registered", n)
		panic(msg)
	}
	Registry[n] = f
}

// GetPlugin returns an instance of the requested plugin
func GetPlugin(n string) (Plugin, bool) {
	f, ok := Registry[n]
	if !ok {
		return nil, false
	}

	return f(), true
}
