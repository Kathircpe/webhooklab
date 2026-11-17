export default function EmptyState({ title, description }: { title: string; description?: string }) {
  return (
    <div className="text-center py-16">
      <div className="text-lg font-medium">{title}</div>
      {description && <div className="text-sm text-gray-500 mt-2">{description}</div>}
    </div>
  );
}
