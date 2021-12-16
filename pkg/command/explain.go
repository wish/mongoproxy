package command

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("explain", func() Command {
		return &Explain{}
	})
}

// the struct for the 'update' command.
type Explain struct {
	Cmd       Command `bson:"explain"`
	Verbosity string  `bson:"verbosity,omitempty"`
	Comment   string  `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *Explain) GetCollection() string {
	return GetCommandCollection(m.Cmd)
}

func (m *Explain) FromBSOND(d bson.D) error {
	type Alias Explain
	aux := &struct {
		Cmd    bson.D `bson:"explain"`
		*Alias `bson:",inline"`
	}{
		Alias: (*Alias)(m),
	}

	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&aux); err != nil {
		return err
	}

	cmd, ok := GetCommand(aux.Cmd[0].Key)
	if !ok {
		return fmt.Errorf("Unable to load command within explain: '" + d[0].Key + "'")
	}

	// Re-pack common into aux.Cmd
	b, err := bson.Marshal(m.Common)
	if err != nil {
		return err
	}

	var commonD bson.D
	if err := bson.Unmarshal(b, &commonD); err != nil {
		return err
	}

	aux.Cmd = append(aux.Cmd, commonD...)

	if err := cmd.FromBSOND(aux.Cmd); err != nil {
		return err
	}

	m.Cmd = cmd

	return nil
}
