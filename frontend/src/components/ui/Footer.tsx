export function Footer() {
  return (
    <footer className="py-4 px-6 bg-white dark:bg-gray-900 border-t border-gray-100 dark:border-gray-800 shrink-0">
      <div className="flex flex-col sm:flex-row items-center justify-between gap-2 text-xs text-gray-400 dark:text-gray-500">
        <span>© {new Date().getFullYear()} Pingr. All rights reserved.</span>
        <div className="flex items-center gap-4">
          <a href="/privacy" className="hover:text-gray-600 dark:hover:text-gray-300 transition-colors">Privacy Policy</a>
          <a href="/terms" className="hover:text-gray-600 dark:hover:text-gray-300 transition-colors">Terms of Use</a>
        </div>
      </div>
    </footer>
  )
}
