package tuyacloud

// Credentials identify a Tuya cloud project. Region is one of "us", "eu", "cn" or "in".
type Credentials struct {
	AccessID     string
	AccessSecret string
	Region       string
}
