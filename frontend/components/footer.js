class CarebedFooter extends HTMLElement {
    connectedCallback() {
        const year = new Date().getFullYear();
        this.innerHTML = `
            <footer class="mt-8 text-center pb-6">
                <p class="text-xs text-slate-400 font-medium tracking-wide">
                    &copy; ${year} Carebed. All rights reserved.
                </p>
                <p class="text-[11px] text-slate-400/70 mt-1.5 uppercase tracking-wider">
                    Computer Engineering Thesis Project
                </p>
            </footer>
        `;
    }
}
// Register the custom tag
customElements.define('care-footer', CarebedFooter);