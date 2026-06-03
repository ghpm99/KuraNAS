module.exports = [
  {
    ignores: ['node_modules/**']
  },
  {
    files: ['**/*.js'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        AbortController: 'readonly',
        Blob: 'readonly',
        cancelAnimationFrame: 'readonly',
        URL: 'readonly',
        URLSearchParams: 'readonly',
        chrome: 'readonly',
        clearInterval: 'readonly',
        clearTimeout: 'readonly',
        console: 'readonly',
        CustomEvent: 'readonly',
        DOMParser: 'readonly',
        document: 'readonly',
        Event: 'readonly',
        fetch: 'readonly',
        FormData: 'readonly',
        Headers: 'readonly',
        history: 'readonly',
        location: 'readonly',
        MutationObserver: 'readonly',
        MediaRecorder: 'readonly',
        MediaSource: 'readonly',
        module: 'readonly',
        navigator: 'readonly',
        performance: 'readonly',
        prompt: 'readonly',
        require: 'readonly',
        requestAnimationFrame: 'readonly',
        Response: 'readonly',
        setInterval: 'readonly',
        setTimeout: 'readonly',
        window: 'readonly',
        __dirname: 'readonly'
      }
    },
    rules: {
      'no-undef': 'error'
    }
  }
];
