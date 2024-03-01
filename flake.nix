{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";

    flake-utils = { url = "github:numtide/flake-utils"; };

    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, pre-commit-hooks, nixpkgs, flake-utils, treefmt-nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        version = "dev";

        pkgs = (import nixpkgs) { inherit system; };

        # Eval the treefmt modules from ./treefmt.nix
        treefmtEval = treefmt-nix.lib.evalModule pkgs ./treefmt.nix;
      in {
        # for `nix fmt`
        formatter = treefmtEval.config.build.wrapper;

        # for `nix flake check`
        checks = {
          formatting = treefmtEval.config.build.check self;
          build = self.packages.${system}.default;
          pre-commit-check = let
            # some treefmt formatters are not supported in pre-commit-hooks we
            # filter them out for now.
            toFilter = [ "yamlfmt" ];
            filterFn = n: _v: (!builtins.elem n toFilter);
            treefmtFormatters =
              pkgs.lib.mapAttrs (_n: v: { inherit (v) enable; })
              (pkgs.lib.filterAttrs filterFn (import ./treefmt.nix).programs);
          in pre-commit-hooks.lib.${system}.run {
            src = ./.;
            hooks = treefmtFormatters;
          };
        };

        # Acessible through 'nix develop' or 'nix-shell' (legacy)
        devShells.default = pkgs.mkShell {
          inherit (self.checks.${system}.pre-commit-check) shellHook;
          inputsFrom = [ self.packages.${system}.default ];
        };

        packages = rec {
          default = kubernixos;
          kubernixos = pkgs.buildGoModule rec {
            name = "kubernixos-${version}";
            inherit version;

            src = pkgs.nix-gitignore.gitignoreSource [ ] ./.;

            preBuild = ''
              ldflags+=" -X github.com/dbcdk/kubernixos/nix.root=$out/lib"
            '';

            vendorHash = "sha256-yaVpYhAfddW0INS+2lpjE5lYwo5K82qv74bM9WYAsGs=";

            postInstall = ''
              cp -rv $src/lib $out
            '';

            meta = {
              homepage = "https://github.com/dbcdk/kubernixos";
              description =
                "Kubernixos is a k8s object reconciler written in Golang.";
            };
          };
        };
      });
}
