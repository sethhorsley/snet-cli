class Snet < Formula
  desc 'Secure HTTPS tunnels from localhost to public URLs'
  homepage 'https://github.com/sethhorsley/snet-cli'
  version '0.1.0'
  license 'MIT'

  on_macos do
    if Hardware::CPU.arm?
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.1.0/snet-darwin-arm64.tar.gz'
      sha256 '79cc22bfc82ef62e74496fb32a251ba4c0ccfe62ec32d99e0a62c678cbfa760c'
    else
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.1.0/snet-darwin-amd64.tar.gz'
      sha256 'd8c558139ba70d1861e150a7ca52142bbe22d8f7d15ae2264c5405eb80e4fa34'
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.1.0/snet-linux-arm64.tar.gz'
      sha256 '6d4a72ebaf8abdefb025f296d105951d688ec30933754ac41e9a25b5ed3d9278'
    else
      url 'https://github.com/sethhorsley/snet-cli/releases/download/v0.1.0/snet-linux-amd64.tar.gz'
      sha256 '7dd2c9c197515d24dda7928e5926a52ab51c442166f8aaac1bbd1fc9e0c3a222'
    end
  end

  def install
    bin.install 'snet'
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/snet version")
  end
end
