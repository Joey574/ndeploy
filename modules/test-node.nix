{ ... }:

{
  services.openssh = {
    enable = true;
  };

  nix.settings.trusted-users = [ "root" "admin" ];

  users.users.admin = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    openssh.authorizedKeys.keys = [
      "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ7M/4N/2mWT6hBWs9vkiRV/uixjfnYOG3MPhxk5/UHT testkey"
    ];
  };

  system.stateVersion = "26.05";
}
