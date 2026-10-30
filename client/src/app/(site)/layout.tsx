import { Footer } from "@/components/layout/footer";
import { Navbar } from "@/components/layout/navbar";

export default function SiteLayout({ children }: LayoutProps<"/">) {
  return (
    <>
      <a
        href="#content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:rounded-md focus:bg-background focus:px-3 focus:py-2"
      >
        Skip to content
      </a>
      <Navbar />
      <main id="content" className="mx-auto w-full max-w-5xl flex-1 px-4">
        {children}
      </main>
      <Footer />
    </>
  );
}
