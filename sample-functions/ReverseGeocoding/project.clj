(defproject reversegeocoding "0.1.0-SNAPSHOT"
  :description "Get address from GPS location"
  :url "http://example.com/FIXME"
  :license {:name "Eclipse Public License"
            :url "http://www.eclipse.org/legal/epl-v10.html"}
  :dependencies [[org.clojure/clojure "1.8.0"]
                 [org.clojure/tools.logging "0.3.1"]
                 [cheshire "5.5.0"]
                 [clj-http "2.1.0"]]
  :main ^:skip-aot reversegeocoding.core
  :target-path "target/%s"
  :profiles {:uberjar {:aot :all}})
