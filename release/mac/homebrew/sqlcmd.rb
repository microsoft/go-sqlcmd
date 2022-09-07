class Sqlcmd < Formula
  desc "Microsoft SQL Server command-line interface"
  homepage "https://github.com/microsoft/go-sqlcmd"
  url "https://github.com/microsoft/go-sqlcmd.git",
  tag:      "v0.8.1",
  revision: "4e0b95ce49b8164c6496ca1b7a85fa57734fef4c"
  license "MIT"

  depends_on "go" => :build

  def check_eula_acceptance?
    if ENV["ACCEPT_EULA"] != "y" && ENV["ACCEPT_EULA"] != "Y"
      puts "The license terms for this product can be downloaded from"
      puts "https://github.com/microsoft/go-sqlcmd/blob/main/LICENSE."
      puts "By entering 'YES', you indicate that you accept the license terms."
      puts ""
      loop do
        puts "Do you accept the license terms? (Enter YES or NO)"
        accept_eula = $stdin.gets.chomp
        if accept_eula
          break if accept_eula.casecmp("YES").zero?

          if accept_eula.casecmp("NO").zero?
            puts "Installation terminated: License terms not accepted."
            return false
          else
            puts "Please enter YES or NO"
          end
        else
          puts "Installation terminated: Could not prompt for license acceptance."
          puts "If you are performing an unattended installation, you may set"
          puts "ACCEPT_EULA to Y to indicate your acceptance of the license terms."
          return false
        end
      end
    end
    true
  end

  def install
    return false unless check_eula_acceptance?

    ENV["CGO_ENABLED"] = "0"

    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/sqlcmd"
  end

  test do
    out = shell_output("#{bin}/sqlcmd -?")
    assert_match "Usage: sqlcmd", out
  end
end
