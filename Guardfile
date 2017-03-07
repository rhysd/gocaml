def run_test(file)
  dir = file.match(%r[^[^/]+])[0]
  result = `go test -v ./#{file} #{Dir["./#{dir}/*.go"].reject{|p| p.end_with? '_test.go'}.join(' ')}`
  puts result.gsub(/\bRUN\b/, "\e[1;93mRUN\e[0m").gsub(/\bPASS\b/, "\e[1;92mPASS\e[0m").gsub(/\bFAIL\b/, "\e[1;91mFAIL\e[0m")
end

def sep(f)
  puts "\033[93m#{Time.now}: #{File.basename f}\033[0m"
end

guard :shell do
  watch /\.go$/ do |m|
    case m[0]
    when /_test\.go$/
      sep m[0]
      # dir = m[0].match(%r[^[^/]+])[0]
      # system "go test ./#{dir}"
      run_test m[0]
    else
      system "make build"
    end
  end
  watch /(.*)\/testdata\/.+\.(:?ml|out)$/ do |m|
    sep m[0]
    system "go test ./#{m[1]}"
  end
  watch /\.c$/ do |m|
    system "make build"
  end
  watch /\.go\.y$/ do |m|
    system "make build"
  end
end
