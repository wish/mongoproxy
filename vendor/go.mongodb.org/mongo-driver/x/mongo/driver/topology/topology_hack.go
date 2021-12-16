package topology

func (t *Topology) ProcessSRVResults(parsedHosts []string) bool {
	return t.processSRVResults(parsedHosts)
}
