// File: shared/utils.mjs

/**
 * Dynamically loads a CSS file if it hasn't been loaded already.
 * @param {string} path - The path to the CSS file.
 */
export function loadCSS(path) {
    if (!document.querySelector(`link[href="${path}"]`)) {
        const link = document.createElement('link');
        link.rel = 'stylesheet';
        link.href = path;
        document.head.appendChild(link);
    }
}

/**
 * Dynamically loads a JavaScript file from a URL if it hasn't been loaded already.
 * Returns a promise that resolves when the script is loaded.
 * @param {string} url - The URL of the script to load.
 * @returns {Promise<void>}
 */
export function loadScript(url) {
    return new Promise((resolve, reject) => {
        if (document.querySelector(`script[src="${url}"]`)) {
            resolve(); // Already loaded
            return;
        }
        const script = document.createElement('script');
        script.src = url;
        script.onload = () => resolve();
        script.onerror = () => reject(new Error(`Failed to load script: ${url}`));
        document.head.appendChild(script);
    });
}