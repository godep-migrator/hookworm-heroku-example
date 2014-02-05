#!/usr/bin/env ruby
require 'hookworm-handlers'

exit Hookworm::LoggingHandler.new.run!(ARGV)
