{ pkgs ? (import <nixpkgs> {})
, version ? "dev"
}:

pkgs.buildGoModule rec {
  name = "kubernixos";
  src = pkgs.nix-gitignore.gitignoreSource [] ./.;

  preBuild = ''
    ldflags+=" -X github.com/dbcdk/kubernixos/nix.root=$out/lib"
  ''; 

  vendorSha256 = "sha256-d8PZ3RDd4c3kXxd9dwFX0ceHiDBtOYCNb19wY4yiUDg=";

  postInstall = ''
    cp -rv $src/lib $out
  '';

  meta = {
    homepage = "https://github.com/dbcdk/kubernixos";
    description = "Kubernixos is a k8s object reconciler written in Golang.";
  };
}
