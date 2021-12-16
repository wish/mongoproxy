package command

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestReadPref(t *testing.T) {
	in := bson.D{
		{"find", "coll"},
		{"$readPreference", bson.D{{"mode", "secondary"}}},
	}

	cmd, _ := GetCommand(in[0].Key)
	if err := cmd.FromBSOND(in); err != nil {
		t.Fatal(err)
	}

	findCmd := cmd.(*Find)

	fmt.Println(findCmd.ReadPreference)
	if GetCommandReadPreferenceMode(cmd) != "secondary" {
		t.Fatalf("Mismatch expected=secondary actual=%s", GetCommandReadPreferenceMode(cmd))
	}
}
