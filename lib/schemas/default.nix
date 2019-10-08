{ stdenv, fetchFromGitHub, buildGoPackage }:

stdenv.mkDerivation rec {
  name = "kubernetes-schemas-${version}";
  version = "6a498a60dc68c5f6a1cc248f94b5cd1e7241d699";

  src = fetchFromGitHub {
    owner = "instrumenta";
    repo = "kubernetes-json-schema";
    rev = version;
    sha256 = "1y9m2ma3n4h7sf2lg788vjw6pkfyi0fa7gzc870faqv326n6x2jr";
  };

  phases = [ "installPhase" ];

  installPhase = ''
    mkdir -p $out/kubernetes-json-schema
    cp -r $src $out/kubernetes-json-schema/master
  '';

  meta = {
    homepage = "https://github.com/instrumenta/kubernetes-json-schema";
    description = ''
      A set of JSON schemas for various Kubernetes versions,
      extracted from the OpenAPI definitions.
    '';
  };
}
