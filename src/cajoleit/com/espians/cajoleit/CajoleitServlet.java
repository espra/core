package com.espians.cajoleit;

import java.io.IOException;
import javax.servlet.http.*;

public class CajoleitServlet extends HttpServlet {
	public void doGet(HttpServletRequest req, HttpServletResponse resp)
			throws IOException {
		resp.setContentType("text/plain");
		resp.getWriter().println("Hello, world!");
	}
}