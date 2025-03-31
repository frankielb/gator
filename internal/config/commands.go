package config

import "fmt"

type State struct {
	CurrentConfig *Config
}

type Command struct {
	Name string
	Args []string
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}
	err := s.CurrentConfig.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}
	fmt.Printf("User %v set\n", cmd.Args[0])
	return nil
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	if c.Handlers == nil {
		c.Handlers = make(map[string]func(*State, Command) error)
	}
	c.Handlers[name] = f
}
func (c *Commands) Run(s *State, cmd Command) error {
	if handler, exists := c.Handlers[cmd.Name]; exists {
		return handler(s, cmd)
	} else {
		return fmt.Errorf("Command not found: %s", cmd.Name)
	}

}
