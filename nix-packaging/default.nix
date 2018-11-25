{ stdenv, fetchgit, buildGoPackage, go-bindata, lib,
  version ? "dev"
}:

with builtins; with lib;

let
    filter = path: type:
    let
        base = baseNameOf(path);
        dir = baseNameOf(dirOf path);
        isGo = hasSuffix ".go" base;
        isNix = hasSuffix ".nix" base;
    in
        (type == "directory" && base == "lib") ||
        (type == "directory" && base == "assets") ||
        (type == "directory" && base == "nix") ||
        (type == "directory" && base == "kubeclient") ||
        (type == "directory" && base == "kubectl") ||
        (type == "regular" && isGo) ||
        (type == "regular" && dir == "lib" && isNix);
in
(buildGoPackage rec {
  name = "kubernixos-unstable-${version}";
  inherit version;

  goPackagePath = "github.com/dbcdk/kubernixos";
  src = filterSource filter ./..;
  goDeps = ./deps.nix;

  outputs = [ "out" "bin" ];

  buildInputs = [ go-bindata ];
  prePatch = ''
    go-bindata -pkg assets -o assets/assets.go lib/
  '';

  meta = {
    homepage = "https://github.com/dbcdk/kubernixos";
    description = "Kubernixos is a k8s object reconciler written in Golang.";
  };
})
