import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

/** Renders a project's markdown body with the site's typography. */
export function ProjectBody({ markdown }: { markdown: string }) {
  return (
    <div className="space-y-4 leading-relaxed">
      <Markdown
        remarkPlugins={[remarkGfm]}
        components={{
          h2: ({ node, ...props }) => (
            <h2
              className="mt-8 text-xl font-semibold tracking-tight"
              {...props}
            />
          ),
          h3: ({ node, ...props }) => (
            <h3 className="mt-6 text-lg font-semibold" {...props} />
          ),
          ul: ({ node, ...props }) => (
            <ul className="list-disc space-y-1 pl-6" {...props} />
          ),
          ol: ({ node, ...props }) => (
            <ol className="list-decimal space-y-1 pl-6" {...props} />
          ),
          a: ({ node, ...props }) => (
            <a
              className="underline underline-offset-4"
              target="_blank"
              rel="noreferrer"
              {...props}
            />
          ),
          code: ({ node, ...props }) => (
            <code
              className="rounded bg-muted px-1 py-0.5 font-mono text-sm"
              {...props}
            />
          ),
        }}
      >
        {markdown}
      </Markdown>
    </div>
  );
}
