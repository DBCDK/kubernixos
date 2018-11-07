 { packages ? <nixpkgs>, modules ? [] }:
let
  option = with pkgs.lib; mkOption {
    type = types.listOf types.attrs;
  };

  pkgs = if builtins.isAttrs packages then packages else (import packages {});

  merge = item: pkgs.lib.recursiveUpdate item {
    metadata = {
      labels = {
        reconciler = "kubernixos";
      };
    };
  };
in
{
  manifests = {
    apiVersion = "v1";
    kind = "List";
    items = map merge ((import "${toString pkgs.path}/nixos/lib/eval-config.nix" {
     inherit pkgs;
     modules = modules ++ [ ({ options.kubernixos.manifests = option; }) ];
    }).config.kubernixos.manifests);
 };
}
