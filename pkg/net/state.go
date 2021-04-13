package net

type State struct {
	VethPairs  *ActiveVethPairs  `yaml:"veth_pairs"`
	Namespaces *ActiveNamespaces `yaml:"namespaces"`
}

//func CreateState(yaml string) (*State, error) {
//
//}

//func (s *State) ToYaml() (string, error) {
//
//}
