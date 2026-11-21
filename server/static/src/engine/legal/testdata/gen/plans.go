package main

type plansFile struct {
	fixtures map[string]exportedFixture
	games    map[string]exportedGame
}

func (p plansFile) sections() []section {
	return []section{{key: "comment", scalar: generatedComment}}
}

func buildPlans() plansFile { return plansFile{} }

type evaluableFile struct {
	fixtures map[string]exportedFixture
	games    map[string]exportedGame
}

func (e evaluableFile) sections() []section {
	return []section{{key: "comment", scalar: generatedComment}}
}

func buildEvaluable() evaluableFile { return evaluableFile{} }
