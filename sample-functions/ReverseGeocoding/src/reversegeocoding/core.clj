(ns reversegeocoding.core
  (:gen-class)
  (:require [clj-http.client :as www])
  (:require [clojure.tools.logging :as log])
  (:require [clojure.string :as str])
  (:require [cheshire.core :refer :all])
  (:use [slingshot.slingshot :only [throw+ try+]]))

(defn -main
  "Get address from geolocation coordinates"
  [& args]
  (try+
    (let [input (read-line)
          coords (str/split input #" ")
          lat (second coords)
          lng (first coords)
          data (www/get (format "http://maps.googleapis.com/maps/api/geocode/json?latlng=%s,%s" lat lng) {:accept :json})
          info (parse-string (:body data) true)
          address (:formatted_address (first (:results info)))]
          (println address))
    (catch Object e
      (println (:body e))
      (println (:error_message (parse-string (:body e) true))))))
