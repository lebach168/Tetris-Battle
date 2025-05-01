export default function SidePanel() {
    return (
      <div className="flex flex-col gap-4 text-white">
        <div className="p-2 border-2 border-gray-500">
          <h2 className="text-center text-lg mb-2">SCORE</h2>
          <div className="text-center text-2xl font-bold">000000</div>
        </div>
  
        <div className="p-2 border-2 border-gray-500">
          <h2 className="text-center text-lg mb-2">NEXT</h2>
          <div className="bg-black w-24 h-24 mx-auto flex items-center justify-center">
            {/* Block next sẽ render ở đây */}
          </div>
        </div>
  
        <div className="p-2 border-2 border-gray-500">
          <h2 className="text-center text-lg mb-2">LEVEL</h2>
          <div className="text-center text-2xl font-bold">01</div>
        </div>
      </div>
    );
  }