package repositories

type dockerRegistry struct{}

func NewDockerRegistry() (r *dockerRegistry, e error) {
	r = &dockerRegistry{}

	return
}

func (r *dockerRegistry) Destroy() (e error) {
	return
}

func (r *dockerRegistry) BuildAndPushServerImage(statusCode int) (e error) {
	return
}

func (r *dockerRegistry) BuildAndPushAlduinImage() (e error) {
	return
}
