// File: components/view-switcher/view-switcher.mjs
// Responsibility: Handle view switching via dropdown selector

/**
 * Initialize view switcher (dropdown-based navigation)
 * Specs: Section 4.3 (modified for dropdown instead of tabs)
 */
export function initViewSwitcher() {
    const dropdown = document.getElementById('view-dropdown');
    const views = document.querySelectorAll('.app-view');
    const editorStatus = document.getElementById('editor-status');

    if (!dropdown) {
        console.error('View dropdown not found');
        return;
    }

    dropdown.addEventListener('change', (e) => {
        const targetView = e.target.value;

        // Update active view
        views.forEach(v => v.classList.remove('active'));
        const viewElement = document.getElementById(`${targetView}-view`);

        if (viewElement) {
            viewElement.classList.add('active');

            // Show/hide editor status bar (connection status is always visible)
            if (editorStatus) {
                if (targetView === 'editor') {
                    editorStatus.classList.add('active');
                } else {
                    editorStatus.classList.remove('active');
                }
            }

            // Dispatch view changed event
            document.dispatchEvent(new CustomEvent('view:changed', {
                detail: { view: targetView }
            }));
        }
    });
}
