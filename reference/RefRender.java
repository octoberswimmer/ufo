import java.io.*;
import java.nio.file.*;
import java.util.*;
import org.xhtmlrenderer.pdf.ITextRenderer;
import org.xhtmlrenderer.render.*;
import org.xhtmlrenderer.layout.*;

/** Renders an XHTML file with Flying Saucer and writes the PDF plus a dump of the laid-out box tree. */
public class RefRender {
	public static void main(String[] args) throws Exception {
		File in = new File(args[0]);
		ITextRenderer renderer = ITextRenderer.fromUrl(in.toURI().toURL().toString());
		renderer.layout();
		try (OutputStream os = new FileOutputStream(args[1])) {
			renderer.createPDF(os);
		}
		if (args.length > 2) {
			StringBuilder sb = new StringBuilder();
			dump(renderer.getRootBox(), 0, sb);
			List<PageBox> pages = renderer.getRootBox().getLayer().getPages();
			for (int i = 0; i < pages.size(); i++) {
				PageBox p = pages.get(i);
				sb.append("page ").append(i).append(" top=").append(p.getTop()).append(" bottom=").append(p.getBottom()).append("\n");
			}
			Files.writeString(Path.of(args[2]), sb.toString());
		}
	}

	static void dump(Box box, int depth, StringBuilder sb) {
		sb.append("  ".repeat(depth)).append(box.getClass().getSimpleName());
		if (box.getElement() != null) sb.append(" <").append(box.getElement().getNodeName()).append(">");
		sb.append(" abs=").append(box.getAbsX()).append(",").append(box.getAbsY())
			.append(" size=").append(box.getWidth()).append("x").append(box.getHeight()).append("\n");
		if (box instanceof InlineLayoutBox ilb) {
			for (Object child : ilb.getInlineChildren()) {
				if (child instanceof InlineText t) {
					sb.append("  ".repeat(depth + 1)).append("text x=").append(t.getX()).append(" w=").append(t.getWidth())
						.append(" '").append(t.getSubstring()).append("'\n");
				} else if (child instanceof Box b) {
					dump(b, depth + 1, sb);
				}
			}
			return;
		}
		for (Box child : box.getChildren()) dump(child, depth + 1, sb);
	}
}
