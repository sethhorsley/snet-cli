class Snet < Formula
  desc 'Secure HTTPS tunnels from localhost to public URLs'
  homepage 'https://github.com/sethhorsley/snet-cli'
  version '0.0.0'
  license 'MIT'

  on_macos do
    if Hardware::CPU.arm?
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.0.0/snet-darwin-arm64.tar.gz'
      sha256 '0000000000000000000000000000000000000000000000000000000000000000'
    else
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.0.0/snet-darwin-amd64.tar.gz'
      sha256 '0000000000000000000000000000000000000000000000000000000000000000'
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.0.0/snet-linux-arm64.tar.gz'
      sha256 '0000000000000000000000000000000000000000000000000000000000000000'
    else
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.0.0/snet-linux-amd64.tar.gz'
      sha256 '0000000000000000000000000000000000000000000000000000000000000000'
    end
  end

  def install
    bin.install 'snet'
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/snet version")
  end
end
