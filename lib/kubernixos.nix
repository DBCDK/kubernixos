 { packages ? <nixpkgs>, modules ? [] }:
let
  pkgs = if builtins.isAttrs packages then packages else (import packages {});

  merge = with pkgs; with builtins; cfg: name: item:
  let
    kubeval = pkgs.callPackage ./kubeval {};
    schemas = pkgs.callPackage ./schemas {};

    enriched = lib.recursiveUpdate item {
      metadata = {
        labels = {
          reconciler = "kubernixos";
        };
      };
    };

    validate = runCommand "validate-${name}" {} ''
      mkdir -p $out
      echo '${toJSON enriched}' | \
        ${kubeval}/bin/kubeval --strict -v ${cfg.version} \
        --filename=${name} \
        --schema-location=file://${schemas}

        echo -n "true" >$out/result
    '';
  in
    if (builtins.readFile "${validate}/result") == "true"
    then enriched
    else throw "${name} has validation errors";

in
{
  manifests = with pkgs;
  let
    cfg = (import "${toString path}/nixos/lib/eval-config.nix" {
     inherit pkgs;
     modules = modules ++ [ (import ./module.nix) ];
    }).config.kubernixos;
  in
  {
    apiVersion = "v1";
    kind = "List";
    items = lib.mapAttrsToList (merge cfg) cfg.manifests;
  };
}
