// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package com.espians.cajoleit;

import java.io.IOException;
import javax.servlet.http.*;

import org.mozilla.javascript.JavaScriptException;

// -----------------------------------------------------------------------------
// Servlet
// -----------------------------------------------------------------------------

public class CajoleitServlet extends HttpServlet {

	private static CoffeeScript compiler = null;

	public static CoffeeScript getCompiler() {
		if (compiler == null) {
			compiler = new CoffeeScript();
		}
		return compiler;
	}

	public void doGet(HttpServletRequest req, HttpServletResponse resp)
			throws IOException {
		resp.setContentType("text/plain");
		resp.getWriter().println("Hello, world!");
	}

	public void doPost(HttpServletRequest req, HttpServletResponse resp)
		throws IOException {

		String source = req.getParameter("source");
		String type = req.getParameter("type");
		String compiledCoffee;

		try {
			compiledCoffee = getCompiler().compile(source);
		} catch (JavaScriptException e) {
			resp.getWriter().println(String.format("{\"error\": \"%s\"}", e.getValue().toString()));
			return;
		}

		resp.setContentType("text/plain");
		resp.getWriter().println(compiledCoffee);
	}

}