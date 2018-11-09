{ stdenv, fetchFromGitHub, buildGoPackage }:

stdenv.mkDerivation rec {
  name = "kubernetes-schemas-${version}";
  version = "c7672fd48e1421f0060dd54b6620baa2ab7224ba";

  src = fetchFromGitHub {
    owner = "garethr";
    repo = "kubernetes-json-schema";
    rev = version;
    sha256 = "0picr3wvjx4qv158jy4f60pl225rm4mh0l97pf8nqi9h9x4x888p";
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
