{ nixpkgs ? import ./nixpkgs.nix, pkgs ? import nixpkgs { }, version, }:

pkgs.buildGoModule rec {
  name = "kubernixos-${version}";
  inherit version;

  src = pkgs.nix-gitignore.gitignoreSource [ ] ./.;

  preBuild = ''
    ldflags+=" -X github.com/dbcdk/kubernixos/nix.root=$out/lib"
  '';

  vendorHash = "sha256-vBWyfDW//7CCsSEp1xJlgR41lMhSIJiReJ0wTq2BTzs=";

  postInstall = ''
    cp -rv $src/lib $out
  '';

  meta = {
    homepage = "https://github.com/dbcdk/kubernixos";
    description = "Kubernixos is a k8s object reconciler written in Golang.";
  };
}
