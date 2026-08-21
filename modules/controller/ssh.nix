{ pkgs, ... }:

{
  programs.ssh = {
    enable = true;
    startAgent = true;
    askPassword = "${pkgs.x11_ssh_askpass}/libexec/x11-ssh-askpass";

    extraConfig = ''
      AddKeysToAgent yes
    '';
  };

  systemd.tmpfiles.rules = [
    "d /home/controller/.ssh 0700 controller users -"
    "C /home/controller/.ssh/id_ed25519 0600 controller users - ${../../keys/test_key}"
  ];
}
