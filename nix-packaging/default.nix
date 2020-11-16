{ stdenv, fetchgit, buildGoPackage, go-bindata, lib, removeReferencesTo, go,
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
        (type == "directory" && base == "kubeval") ||
        (type == "directory" && base == "schemas") ||
        (type == "directory" && base == "assets") ||
        (type == "directory" && base == "nix") ||
        (type == "directory" && base == "kubeclient") ||
        (type == "directory" && base == "kubectl") ||
        (type == "regular" && isGo) ||
        (type == "regular" && isNix);
in
(buildGoPackage rec {
  name = "kubernixos-unstable-${version}";
  inherit version;

  goPackagePath = "github.com/dbcdk/kubernixos";
  src = filterSource filter ./..;
  goDeps = ./deps.nix;

  nativeBuildInputs = [ go-bindata removeReferencesTo ];
  prePatch = ''
    go-bindata -pkg assets -o assets/assets.go lib/
  '';

  postInstall = ''
    cp -rv $src/lib $out/
    find $out/bin -type f -exec remove-references-to -t ${go} '{}' +
  '';

  meta = {
    homepage = "https://github.com/dbcdk/kubernixos";
    description = "Kubernixos is a k8s object reconciler written in Golang.";
  };
})
