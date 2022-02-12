package clusters

type kindCluster struct{}

func NewKindCluster() (c *kindCluster, e error) {
	c = &kindCluster{}

	return
}

func (c *kindCluster) Destroy() (e error) {
	return
}

func (c *kindCluster) DeployServer() (e error) {
	return
}

func (c *kindCluster) DeployAlduin() (e error) {
	return
}
