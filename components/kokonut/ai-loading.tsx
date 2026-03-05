'use client'

export function AILoading() {
  return (
    <div className="flex items-center justify-center w-full h-full">
      <div className="flex flex-col items-center gap-4">
        <div className="relative w-12 h-12">
          <div className="absolute inset-0 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full opacity-20 animate-pulse" />
          <div className="absolute inset-0 animate-spin">
            <div className="absolute top-0 left-1/2 w-2 h-2 bg-blue-500 rounded-full -translate-x-1/2" />
            <div className="absolute top-1/4 right-1/4 w-2 h-2 bg-purple-500 rounded-full" />
            <div className="absolute bottom-1/4 right-1/4 w-2 h-2 bg-blue-400 rounded-full" />
          </div>
        </div>
        <p className="text-sm text-muted-foreground">Loading...</p>
      </div>
    </div>
  )
}
