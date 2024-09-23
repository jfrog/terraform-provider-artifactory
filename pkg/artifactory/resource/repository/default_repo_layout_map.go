package repository

// Consolidated list of Default Repo Layout for all Package Types with active Repo Types
var defaultRepoLayoutMap = map[string]SupportedRepoClasses{
	AlpinePackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	AnsiblePackageType: {
		RepoLayoutRef: "ansible-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	BowerPackageType: {
		RepoLayoutRef: "bower-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	CranPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	CargoPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"federated": true,
		},
	},
	ChefPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	CocoapodsPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	ComposerPackageType: {
		RepoLayoutRef: "composer-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	ConanPackageType: {
		RepoLayoutRef: "conan-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	CondaPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	DebianPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	DockerPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	GemsPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	GenericPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	GitLFSPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	GoPackageType: {
		RepoLayoutRef: "go-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	GradlePackageType: {
		RepoLayoutRef: "maven-2-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	HelmPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	HelmOCIPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	HuggingFacePackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   false,
			"federated": false,
		},
	},
	IvyPackageType: {
		RepoLayoutRef: "ivy-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	MavenPackageType: {
		RepoLayoutRef: "maven-2-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	NPMPackageType: {
		RepoLayoutRef: "npm-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	NugetPackageType: {
		RepoLayoutRef: "nuget-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	OCIPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	OpkgPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	P2PackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"remote":  true,
			"virtual": true,
		},
	},
	PubPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	PuppetPackageType: {
		RepoLayoutRef: "puppet-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	PyPiPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	RPMPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	SwiftPackageType: {
		RepoLayoutRef: "swift-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	SBTPackageType: {
		RepoLayoutRef: "sbt-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	TerraformPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     false,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	TerraformBackendPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    false,
			"virtual":   false,
			"federated": false,
		},
	},
	TerraformModulePackageType: {
		RepoLayoutRef: "terraform-module-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	TerraformProviderPackageType: {
		RepoLayoutRef: "terraform-provider-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"remote":    true,
			"virtual":   true,
			"federated": true,
		},
	},
	VagrantPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"local":     true,
			"federated": true,
		},
	},
	VCSPackageType: {
		RepoLayoutRef: "simple-default",
		SupportedRepoTypes: map[string]bool{
			"remote": true,
		},
	},
}
