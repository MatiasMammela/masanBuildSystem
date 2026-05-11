;;; mbs.el --- Major mode for MBS lua build files
(require 'lua-mode)
(require 'flycheck)

(defvar mbs-functions
  '((project      1
                  "project(name string, build_dir_path string) *Project
Initializes a new project with the specified name.
build_dir_path is optional and will default to build/")

    (version      1
                  "version(version float64) void
Enforce a minimum version from the users mbs build system.")

    (build        1
                  "build(project *Project) void
Builds project.")

    (debug        1
                  "debug(project *Project) void
Prints project.")

    (copy         2
              "copy(src string|*File|*Directory..., dest_path string) void
Copy files or directories to dest.
Accepts strings, glob_files or glob_dirs results.
Note: Wildcards and ~/ paths are not supported.")

    (glob_files   1
                  "glob_files(path string...) *File
Globs files with the given path.")

    (glob_dirs    1
                  "glob_dirs(path string...) *Directory
Globs directories with the given path.")

    (glob_packages 1
                   "glob_packages(pkg_name string...) *Package
Globs packages using pkg-config. If not found, tries to install via package manager.")

    (sources      2
                  "sources(project *Project, sources *Files...) void
Binds sources to project.")

    (headers      2
                  "headers(project *Project, headers *Directory...) void
Binds headers to project.")

    (packages     2
                  "packages(project *Project, package *Package...) void
Binds packages to project.")

    (compiler     2
                  "compiler(project *Project, compiler string) void
Binds compiler to project.")

    (assembler    2
                  "assembler(project *Project, assembler string) void
Binds assembler to project.")

    (cflags       2
                  "cflags(project *Project, flag string...) void
Binds cflags to project.")

    (lflags       2
                  "lflags(project *Project, flag string...) void
Binds lflags to project.")

    (asmflags     2
                  "asmflags(project *Project, flag string...) void
Binds asm flags to project.")

    (linkerflags  2
                  "linkerflags(project *Project, flag string...) void
Binds linker flags to project.")

    (autoconfigure 2
                   "autoconfigure(project *Project, enabled bool) void
Sets autoconfigure on or off. Enabled by default. Runs with build if enabled."))

  "MBS function definitions: (name min-arg-count docstring)")

(defun mbs--names ()
  (mapcar (lambda (f) (symbol-name (car f))) mbs-functions))

(defun mbs--doc (name)
  (caddr (assq (intern name) mbs-functions)))

(defun mbs--arg-count (name)
  (cadr (assq (intern name) mbs-functions)))

(defun mbs--signature (name)
  "Return only the first line of doc for NAME."
  (let ((doc (mbs--doc name)))
    (when doc
      (car (split-string doc "\n")))))

(defun mbs-find-variable-name ()
  "Find the variable name assigned to require('mbs')."
  (save-excursion
    (goto-char (point-min))
    (when (re-search-forward
           "\\([a-zA-Z_][a-zA-Z0-9_]*\\)\\s-*=\\s-*require\\s-*(\\s-*[\"']mbs[\"']\\s-*)"
           nil t)
      (match-string 1))))

(defun mbs-match-function (limit)
  "Match mbs function calls dynamically."
  (let* ((var (or (mbs-find-variable-name) "mbs"))
         (pattern (concat "\\b" (regexp-quote var) "\\.[a-zA-Z_]+")))
    (re-search-forward pattern limit t)))

(defun mbs-font-lock-keywords ()
  (list (list #'mbs-match-function 0 font-lock-function-name-face)))





(defun mbs-colorize-doc (doc)
  "Add colors to mbs function signature."
  (when doc
    (with-temp-buffer
      (insert doc)
      (goto-char (point-min))
      (let ((sig-end (save-excursion
                       (end-of-line)
                       (point))))
        ;; Colorize function name before (
        (when (re-search-forward "^\\([a-zA-Z_]+\\)(" sig-end t)
          (put-text-property (match-beginning 1) (match-end 1)
                             'face font-lock-function-name-face))
        ;; Colorize type names only in signature line
        (goto-char (point-min))
        (while (re-search-forward
                "\\(string\\|float64\\|bool\\|void\\|int\\)\\|\\(\\*[A-Z][a-zA-Z]*\\)"
                sig-end t)
          (put-text-property (match-beginning 0) (match-end 0)
                             'face font-lock-type-face)))
      (buffer-string))))

(defun mbs-eldoc-function ()
  "Return documentation for mbs function at point."
  (save-excursion
    (let* ((var (or (mbs-find-variable-name) "mbs"))
           (prefix (concat var "."))
           (end (progn (skip-chars-forward "a-zA-Z_") (point)))
           (start (progn (skip-chars-backward "a-zA-Z_") (point)))
           (word (buffer-substring-no-properties start end))
           (pre-start (- start (length prefix))))
      (when (and (>= pre-start 0)
                 (string= (buffer-substring-no-properties pre-start start) prefix))
        (mbs-colorize-doc (mbs--doc word))))))

(defun mbs-check-buffer ()
  "Check mbs function calls for correct argument counts."
  (let ((errors '()))
    (save-excursion
      (goto-char (point-min))
      (while (re-search-forward
              "\\([a-zA-Z_]+\\)\\.\\([a-zA-Z_]+\\)(\\([^)]*\\))" nil t)
        (let* ((func (match-string 2))
               (args (match-string 3))
               (arg-count (if (string= args "") 0
                            (1+ (cl-count ?, args))))
               (expected (mbs--arg-count func))
               (line (line-number-at-pos)))
          (when (and expected (< arg-count expected))
            (push (list line
                        (format "%s requires at least %d argument(s): %s"
                                func expected (mbs--signature func)))
                  errors)))))
    errors))

(flycheck-define-generic-checker 'mbs-checker
  "Checker for MBS lua build files."
  :start (lambda (checker callback)
           (funcall callback 'finished
                    (mapcar (lambda (err)
                              (flycheck-error-new-at
                               (car err) nil 'error (cadr err)
                               :checker checker
                               :filename (buffer-file-name)))
                            (mbs-check-buffer))))
  :modes '(mbs-mode))

(add-to-list 'flycheck-checkers 'mbs-checker)

(defvar-local mbs-error-overlays '())

(defun mbs-clear-error-overlays ()
  "Remove all mbs error overlays."
  (mapc #'delete-overlay mbs-error-overlays)
  (setq mbs-error-overlays '()))

(defun mbs-show-error-overlays ()
  "Show mbs errors as inline overlays."
  (mbs-clear-error-overlays)
  (dolist (err (mbs-check-buffer))
    (let ((line (car err))
          (msg (cadr err)))
      (save-excursion
        (goto-char (point-min))
        (forward-line (1- line))
        (end-of-line)
        (let ((ov (make-overlay (point) (point))))
          (overlay-put ov 'after-string
                       (propertize (concat "  !! " msg)
                                   'face '(:foreground "red")))
          (push ov mbs-error-overlays))))))



(define-derived-mode mbs-mode lua-mode "MBS"
  "Major mode for MBS lua build files."
  (font-lock-add-keywords nil (mbs-font-lock-keywords) t)
  (setq-local eldoc-documentation-function #'mbs-eldoc-function)
  (setq-local eldoc-echo-area-use-multiline-p t)
  (setq-local flycheck-display-errors-function nil)
  (add-hook 'after-change-functions
            (lambda (_beg _end _len) (mbs-show-error-overlays))
            nil t)
  (add-hook 'find-file-hook #'mbs-show-error-overlays nil t)
  (mbs-show-error-overlays)
  (eldoc-mode 1)
  (flycheck-mode 1))

(add-to-list 'auto-mode-alist '("build\\.lua\\'" . mbs-mode))

(provide 'mbs)
