{ nixpkgs ? import ./nixpkgs.nix, pkgs ? import nixpkgs { }, version }:

pkgs.buildGoModule rec {
  name = "kubernixos-${version}";
  inherit version;

  src = pkgs.nix-gitignore.gitignoreSource [ ] ./.;

  preBuild = ''
    ldflags+=" -X github.com/dbcdk/kubernixos/nix.root=$out/lib"
  '';

  vendorHash = "sha256-Vrxn6wOKryOd2Wsh6YlIMB6Ke+HExn3TAVU1UuZOdA0=";

  postInstall = ''
    cp -rv $src/lib $out
  '';

  meta = {
    homepage = "https://github.com/dbcdk/kubernixos";
    description = "Kubernixos is a k8s object reconciler written in Golang.";
  };
}
