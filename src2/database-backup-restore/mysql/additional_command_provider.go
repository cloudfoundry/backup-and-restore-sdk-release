package mysql

type AdditionalOptionsProvider interface {
	BuildParams() []string
}

type PurgeGTIDOptionProvider struct{}

func (p PurgeGTIDOptionProvider) BuildParams() []string {
	return []string{"--set-gtid-purged=OFF"}
}

func NewPurgeGTIDOptionProvider() PurgeGTIDOptionProvider {
	return PurgeGTIDOptionProvider{}
}

type EmptyAdditionalOptionsProvider struct{}

func (p EmptyAdditionalOptionsProvider) BuildParams() []string {
	return []string{}
}

func NewEmptyAdditionalOptionsProvider() EmptyAdditionalOptionsProvider {
	return EmptyAdditionalOptionsProvider{}
}
