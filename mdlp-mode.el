;;; mdlp-mode.el --- Markdown Live Preview Minor Mode

;; Copyright (C) 2023 Joshua Marsh

;; Author: Joshua Marsh <joshua.marshian@gmail.com>
;; Homepage: https://github.com/icub3d/mdlp
;; Version: 0.4

;; This file is not part of GNU Emacs.

;; Permission is hereby granted, free of charge, to any person
;; obtaining a copy of this software and associated documentation
;; files (the “Software”), to deal in the Software without
;; restriction, including without limitation the rights to use, copy,
;; modify, merge, publish, distribute, sublicense, and/or sell copies
;; of the Software, and to permit persons to whom the Software is
;; furnished to do so, subject to the following conditions:
;;
;; The above copyright notice and this permission notice shall be
;; included in all copies or substantial portions of the Software.
;;
;; THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
;; EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
;; MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
;; NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
;; BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
;; ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
;; CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
;; SOFTWARE.

;;; Commentary:

;; This package provides a minor mode for markdown files that will
;; preview the markdown in a browser and reload as you save the file.
;;
;; Example installation using use-package and straight.el:
;;
;; (use-package mdlp-mode
;;   :straight (mdlp-mode :type git :host github :repo "icub3d/mdlp")
;;   :hook (markdown-mode . mdlp-mode))

;;; Code:

(defgroup mdlp nil
  "Markdown Live Preview."
  :group 'markdown
  :link '(url-link :tag "Github" "https://github.com/icub3d/mdlp"))

(defcustom mdlp-browser nil
  "The browser to use for previewing markdown. Use default if nil."
  :type 'string
  :group 'mdlp)

(defcustom mdlp-github-token nil
  "The to use for rendering with the GitHub API. This help with rate limiting."
  :type 'string
  :group 'mdlp)

(defcustom mdlp-wait-time 2
  "The number of seconds to wait for the server to come up."
  :type 'integer
  :group 'mdlp)

(defun mdlp--browser (url)
  "Open URL in the browser."
  (if mdlp-browser
	  (let ((browse-url-generic-program mdlp-browser))
		(browse-url-generic url))
	(browse-url url)))

(defvar-local mdlp--process nil
  "The process running the mdlp server.")

(defun mdlp-start-process ()
  "Start the mdlp server."
  (interactive)
  (setq mdlp--process (start-process
   "mdlp-server"
   "*mdlp*"
   "mdlp"
   "-github-token" mdlp-github-token
   "-addr" "localhost:8099"
   buffer-file-name))
  (sleep-for mdlp-wait-time)
  (mdlp--browser "http://localhost:8099"))

(defun mdlp-stop-process ()
  "Start the mdlp server."
  (interactive)
  (when mdlp--process
	(delete-process mdlp--process)
	(kill-buffer "*mdlp*")))

(define-minor-mode mdlp-mode
  "Markdown Live Preview."
  :lighter " MDLP"
  (if mdlp-mode
	  (mdlp-start-process)
	(mdlp-stop-process)))

(provide 'mdlp-mode)
;;; mdlp-mode.el ends here
