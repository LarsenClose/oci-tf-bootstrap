# typed: false
# frozen_string_literal: true

# Homebrew formula for oci-tf-bootstrap
# This is a template - GoReleaser auto-generates the actual formula during releases
#
# Manual installation:
#   brew install larsenclose/tap/oci-tf-bootstrap
#
# Or with this local formula:
#   brew install --build-from-source ./Formula/oci-tf-bootstrap.rb

class OciTfBootstrap < Formula
  desc "Generate ready-to-use Terraform from OCI tenancy discovery"
  homepage "https://github.com/larsenclose/oci-tf-bootstrap"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/larsenclose/oci-tf-bootstrap/releases/download/v#{version}/oci-tf-bootstrap_#{version}_darwin_arm64.tar.gz"
    end
    on_intel do
      url "https://github.com/larsenclose/oci-tf-bootstrap/releases/download/v#{version}/oci-tf-bootstrap_#{version}_darwin_amd64.tar.gz"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/larsenclose/oci-tf-bootstrap/releases/download/v#{version}/oci-tf-bootstrap_#{version}_linux_arm64.tar.gz"
    end
    on_intel do
      url "https://github.com/larsenclose/oci-tf-bootstrap/releases/download/v#{version}/oci-tf-bootstrap_#{version}_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "oci-tf-bootstrap"

    # Install shell completions
    bash_completion.install "completions/oci-tf-bootstrap.bash" => "oci-tf-bootstrap"
    zsh_completion.install "completions/oci-tf-bootstrap.zsh" => "_oci-tf-bootstrap"
    fish_completion.install "completions/oci-tf-bootstrap.fish"
  end

  test do
    assert_match "oci-tf-bootstrap", shell_output("#{bin}/oci-tf-bootstrap --version")
  end
end
