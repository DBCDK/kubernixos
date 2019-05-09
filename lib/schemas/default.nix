{ stdenv, fetchFromGitHub, buildGoPackage }:

stdenv.mkDerivation rec {
  name = "kubernetes-schemas-${version}";
  version = "8aa572595b98d73b2b9415ca576f78e163381b10";

  src = fetchFromGitHub {
    owner = "garethr";
    repo = "kubernetes-json-schema";
    rev = version;
    sha256 = "0fxnx0p9dq5rrbg64d8m13m2w31kkyzbiy064rl3ka46bcawzq2z";
  };

  phases = [ "installPhase" ];

  installPhase = ''
    mkdir -p $out/kubernetes-json-schema
    cp -r $src $out/kubernetes-json-schema/master
  '';

  meta = {
    homepage = "https://github.com/garethr/kubernetes-json-schema";
    description = ''
      A set of JSON schemas for various Kubernetes versions,
      extracted from the OpenAPI definitions.
    '';
  };
}
