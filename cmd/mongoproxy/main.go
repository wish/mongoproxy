package main

// the logic for main is moved into github.com/wish/mongoproxy/pkg/cmd
// to enable easy private plugins while still maintaining static linking.
// This way someone could create a repo which has a couple plugins and a dep
// on this upstream repo and just call the cmd.Main method as we do here.
import (
	// If you are creating your own release, include plugin imports here
	"github.com/wish/mongoproxy/pkg/cmd"
)

func main() {
	cmd.Main()
}
