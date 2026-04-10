export function Footer() {
  return (
    <footer className="border-t border-gray-100 py-6 px-4">
      <div className="max-w-5xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-3 text-sm text-gray-400">
        <span>© {new Date().getFullYear()} Pingr. All rights reserved.</span>
        <div className="flex items-center gap-5">
          <a href="#" className="hover:text-gray-600 transition-colors">Privacy</a>
          <a href="#" className="hover:text-gray-600 transition-colors">Terms</a>
          <a href="#" className="hover:text-gray-600 transition-colors">Status</a>
        </div>
      </div>
    </footer>
  )
}
