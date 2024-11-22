{ lib
, buildGoModule
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "goshot";
  version = "0.4.7"; # Updated to latest version

  src = fetchFromGitHub {
    owner = "watzon";
    repo = "goshot";
    rev = "v${version}";
    hash = "sha256-OJD6J+M8NhkxtY60F6I9lAwfEZnR8jmgdIT8vTbOuSo=";
  };

  # Let Nix handle the vendoring
  vendorHash = "sha256-Nc8aIpbwa0ac2tQUaVVaJHF7XxQcIpCQu5saJnAi3xI=";

  # Disable tests as they require embedded fonts
  doCheck = false;

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "A simple screenshot tool written in Go";
    homepage = "https://github.com/watzon/goshot";
    license = licenses.mit;
    maintainers = with maintainers; [ /* Add your nixpkgs maintainer name here */ ];
    mainProgram = "goshot";
    platforms = platforms.unix;
  };
}
