{ config, pkgs, ... }: {

  options.kubernixos.manifests = with pkgs.lib; mkOption {
    type = types.listOf types.attrs;
  };

}
