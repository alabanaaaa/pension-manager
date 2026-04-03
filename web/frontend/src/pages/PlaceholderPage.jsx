export default function PlaceholderPage({ title, description }) {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight text-neutral-900">{title}</h1>
        <p className="text-neutral-500 mt-1">{description}</p>
      </div>
      <div className="bg-white rounded-2xl border border-neutral-100 p-16 text-center">
        <div className="w-16 h-16 bg-neutral-50 rounded-2xl flex items-center justify-center mx-auto mb-6">
          <svg className="w-8 h-8 text-neutral-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-neutral-900 mb-2">{title} Coming Soon</h3>
        <p className="text-neutral-400 max-w-md mx-auto text-sm">This module is under development. Check back soon for updates.</p>
      </div>
    </div>
  );
}
