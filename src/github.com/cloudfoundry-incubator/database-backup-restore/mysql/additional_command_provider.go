package mysql

type AdditionalOptionsProvider interface {
	BuildParams() []string
}

type DefaultAdditionalOptionsProvider struct{}

func (p DefaultAdditionalOptionsProvider) BuildParams() []string {
	return []string{"--set-gtid-purged=OFF"}
}

func NewDefaultAdditionalOptionsProvider() DefaultAdditionalOptionsProvider {
	return DefaultAdditionalOptionsProvider{}
}

type LegacyAdditionalOptionsProvider struct{}

func (p LegacyAdditionalOptionsProvider) BuildParams() []string {
	return []string{}
}

func NewLegacyAdditionalOptionsProvider() LegacyAdditionalOptionsProvider {
	return LegacyAdditionalOptionsProvider{}
}
