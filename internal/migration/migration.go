package migration

type Runner struct{ Directory string }

func (r Runner) Up() error   { return nil }
func (r Runner) Down() error { return nil }
