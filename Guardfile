def run_test(file)
  dir = file.match(%r[^[^/]+])[0]
  sources = Dir["./#{dir}/*.go"].reject{|p| p.end_with? '_test.go'}.join(' ')
  if file.end_with? 'node_to_type_test.go'
    # XXX
    sources += " ./sema/deref_test.go"
  end
  result = run_tests "./#{file} #{sources}"
  puts_out result
end

def puts_out(out)
  puts out.gsub(/\bRUN\b/, "\e[1;93mRUN\e[0m").gsub(/\bPASS\b/, "\e[1;92mPASS\e[0m").gsub(/\bFAIL\b/, "\e[1;91mFAIL\e[0m")
end

def sep(f)
  puts "\033[93m#{Time.now}: #{File.basename f}\033[0m"
end

def run_tests(args)
 `CGO_LDFLAGS_ALLOW='-Wl,(-search_paths_first|-headerpad_max_install_names)' go test -v #{args}`
end

guard :shell do
  watch /\.go$/ do |m|
    sep m[0]
    case m[0]
    when /_test\.go$/
      run_test m[0]
    else
      system "make build"
    end
  end
  watch /(.*)\/testdata\/.+\.(:?ml|out)$/ do |m|
    sep m[0]
    puts_out run_tests("./#{m[1]}")
  end
  watch /\.c$/ do |m|
    sep m[0]
    system "make build"
  end
  watch /\.go\.y$/ do |m|
    sep m[0]
    system "make build"
  end
end
