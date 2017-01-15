def run_test(file)
  result = `go test -v ./#{file} #{Dir['./*/*.go'].reject{|p| p.end_with? '_test.go'}.join(' ')}`
  puts result.gsub(/\bRUN\b/, "\e[1;93mRUN\e[0m").gsub(/\bPASS\b/, "\e[1;92mPASS\e[0m").gsub(/\bFAIL\b/, "\e[1;91mFAIL\e[0m")
end

guard :shell do
  watch /\.go$/ do |m|
    puts "\033[93m#{Time.now}: #{File.basename m[0]}\033[0m"
    case m[0]
    when /_test\.go$/
      run_test m[0]
    else
      system "go build"
    end
  end
end
