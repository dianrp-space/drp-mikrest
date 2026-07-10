import Swal from 'sweetalert2'

// Konfigurasi default SweetAlert2 dengan tema dark drp-mikrest
const baseConfig = {
  background: '#26344A',
  color: '#D0DAE5',
  confirmButtonColor: '#2563EB',
  cancelButtonColor: '#54677D',
  customClass: {
    popup: 'rounded-xl border border-ink-700 shadow-card',
    title: 'text-ink-100',
    htmlContainer: 'text-ink-300',
    confirmButton: 'rounded-lg px-4 py-2 text-sm font-medium',
    cancelButton: 'rounded-lg px-4 py-2 text-sm font-medium ml-2',
  },
}

// Deteksi apakah string mengandung HTML tag. SweetAlert2 default escape
// `text` (dan `title`), jadi untuk render HTML harus pakai property `html`.
function looksLikeHTML(s: string): boolean {
  return /<[a-z][\s\S]*>/i.test(s)
}

// Helper internal: jika string berisi HTML tag, kirim sebagai `html`; jika tidak,
// kirim sebagai `text` (escape). Return object siap di-spread ke config Swal.
function packContent(content: string | undefined): Record<string, string> {
  if (!content) return {}
  return looksLikeHTML(content) ? { html: content } : { text: content }
}

export function useSwal() {
  const toast = Swal.mixin({
    ...baseConfig,
    toast: true,
    position: 'top-end',
    showConfirmButton: false,
    timer: 3000,
    timerProgressBar: true,
    customClass: {
      ...baseConfig.customClass,
      popup: 'rounded-lg border border-ink-700 shadow-card',
    },
  })

  return {
    Swal,

    // Toast notifikasi singkat
    success: (msg: string) => toast.fire({ icon: 'success', title: msg }),
    error:   (msg: string) => toast.fire({ icon: 'error',   title: msg }),
    info:    (msg: string) => toast.fire({ icon: 'info',    title: msg }),
    warning: (msg: string) => toast.fire({ icon: 'warning', title: msg }),

    // Dialog konfirmasi (return Promise<boolean>)
    confirm: (title: string, content?: string, opts?: Record<string, any>) =>
      Swal.fire({
        ...baseConfig,
        title,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Ya, lanjut',
        cancelButtonText: 'Batal',
        ...packContent(content),
        ...opts,
      }).then((r) => r.isConfirmed),

    // Dialog info dengan tombol OK
    ok: (title: string, content?: string, icon: 'success' | 'error' | 'info' | 'warning' = 'info') =>
      Swal.fire({ ...baseConfig, title, icon, confirmButtonText: 'OK', ...packContent(content) }),

    // Dialog sukses
    successDialog: (title: string, content?: string) =>
      Swal.fire({ ...baseConfig, title, icon: 'success', confirmButtonText: 'OK', ...packContent(content) }),

    // Dialog error
    errorDialog: (title: string, content?: string) =>
      Swal.fire({ ...baseConfig, title, icon: 'error', confirmButtonText: 'OK', ...packContent(content) }),
  }
}
