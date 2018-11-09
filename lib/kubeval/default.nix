{ stdenv, fetchFromGitHub, buildGoPackage }:

buildGoPackage rec {
  name = "kubeval-${version}";
  version = "0.7.3";

  goPackagePath = "github.com/garethr/kubeval";

  src = fetchFromGitHub {
    owner = "garethr";
    repo = "kubeval";
    rev = version;
    sha256 = "042v4mc5p80vmk56wp6aw89yiibjnfqn79c0zcd6y179br4gpfnb";
  };

  goDeps = ./deps.nix;

  meta = {
    homepage = "https://github.com/garethr/kubeval";
    description = "Kubeval is a tool for validating a Kubernetes YAML or JSON configuration file.";
  };
}
