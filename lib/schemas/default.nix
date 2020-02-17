{ stdenv, fetchFromGitHub, buildGoPackage }:

stdenv.mkDerivation rec {
  name = "kubernetes-schemas-${version}";
  version = "74551ef81a466162b63c62038a4b760935e77bf4";

  src = fetchFromGitHub {
    owner = "instrumenta";
    repo = "kubernetes-json-schema";
    rev = version;
    sha256 = "1kxpl9zn7pb6blxi8dpgypv8wk3spadx221iy95n5c447q1jimxw";
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
