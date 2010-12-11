// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package com.espians.cajoleit;

import java.io.IOException;
import javax.servlet.http.*;

import com.google.caja.lexer.CharProducer;
import com.google.caja.lexer.InputSource;
import com.google.caja.lexer.JsLexer;
import com.google.caja.lexer.JsTokenQueue;
import com.google.caja.lexer.ParseException;
import com.google.caja.lexer.TokenConsumer;
import com.google.caja.parser.js.Block;
import com.google.caja.parser.js.CajoledModule;
import com.google.caja.parser.js.Parser;
import com.google.caja.parser.js.UncajoledModule;
import com.google.caja.parser.quasiliteral.ES53Rewriter;
import com.google.caja.parser.quasiliteral.Rewriter;
import com.google.caja.render.Concatenator;
import com.google.caja.render.JsMinimalPrinter;
import com.google.caja.render.JsPrettyPrinter;
import com.google.caja.reporting.BuildInfo;
import com.google.caja.reporting.Message;
import com.google.caja.reporting.MessageLevel;
import com.google.caja.reporting.MessageQueue;
import com.google.caja.reporting.RenderContext;
import com.google.caja.reporting.SimpleMessageQueue;
import com.google.caja.util.Callback;

import org.mozilla.javascript.JavaScriptException;

// -----------------------------------------------------------------------------
// Servlet
// -----------------------------------------------------------------------------

public class CajoleitServlet extends HttpServlet {

	private static CoffeeScript compiler = null;
	private HttpServletResponse resp;

	private static class IOCallback implements Callback<IOException> {
		public IOException exception = null;
		public void handle(IOException e) {
			if (this.exception != null) {
				this.exception = e;
			}
		}
	}

	public static CoffeeScript getCompiler() {
		if (compiler == null) {
			compiler = new CoffeeScript();
		}
		return compiler;
	}

	private void badRequest(String errorMessage)
		throws IOException {
		errorResponse(400, errorMessage);
	}

	private void error(String errorMessage)
		throws IOException {
		errorResponse(500, errorMessage);
	}

	private void errorResponse(Integer statusCode, String errorMessage)
		throws IOException {
		resp.setStatus(statusCode);
		resp.getWriter().println("ERROR: " + errorMessage);
	}

	public void doPost(HttpServletRequest req, HttpServletResponse resp)
		throws IOException {

		this.resp = resp;

		BuildInfo buildInfo;
		CajoledModule cajoledModule;
		Concatenator concatenator;
		CharProducer inputStream;
		MessageQueue messageQueue;
		Block parsedBlock;
		boolean pretty;
		IOCallback renderCallback;
		RenderContext renderContext;
		Rewriter rewriter;
		StringBuilder stringBuilder;
		TokenConsumer tokenConsumer;
		JsTokenQueue tokenQueue;
		UncajoledModule uncajoledModule;

		String source = req.getParameter("source");
		String type = req.getParameter("input_type");
		String prettyParam = req.getParameter("pretty");

		resp.setContentType("text/plain; charset=utf-8");

		if (source == null) {
			badRequest("The `source` parameter was not specified.");
			return;
		}

		if (source.length() > 102400) { // 100kB
			badRequest("The `source` parameter value is too long!");
			return;
		}

		if (type == null) {
			badRequest("The `type` parameter was not specified.");
			return;
		}

		if (type.equals("coffee")) {
			try {
				source = getCompiler().compile(source);
			} catch (JavaScriptException e) {
				error(e.getValue().toString());
				return;
			}
		} else {
			if (!type.equals("js")) {
				badRequest("Unknown `type` parameter value.");
				return;
			}
		}

		cajoledModule = null;
		messageQueue = new SimpleMessageQueue();

		try {
			buildInfo = BuildInfo.getInstance();
			inputStream = CharProducer.Factory.fromString(source, InputSource.PREDEFINED);
			tokenQueue = new JsTokenQueue(new JsLexer(inputStream), InputSource.PREDEFINED);
			parsedBlock = new Parser(tokenQueue, messageQueue).parse();
			tokenQueue.expectEmpty();
			uncajoledModule = new UncajoledModule(parsedBlock);
			rewriter = new ES53Rewriter(buildInfo, messageQueue, false);
			cajoledModule = (CajoledModule) rewriter.expand(uncajoledModule);
		} catch (ParseException e) {
			e.toMessageQueue(messageQueue);
		}

		if (messageQueue.hasMessageAtLevel(MessageLevel.ERROR)) {
			cajoledModule = null;
			String errorMessage = "Couldn't cajole the source.\n\n";
			String inputSourceString = InputSource.PREDEFINED.toString() + ":";
			for (Message m: messageQueue.getMessages()) {
				errorMessage += m.toString().replace(inputSourceString, "Line: ") + "\n\n";
			}
			error(errorMessage);
			return;
		}

		stringBuilder = new StringBuilder();
		renderCallback = new IOCallback();
		concatenator = new Concatenator(stringBuilder, renderCallback);

		if (prettyParam == null) {
			pretty = false;
		} else {
			if (prettyParam.equals("1")) {
				pretty = true;
			} else {
				pretty = false;
			}
		}

		if (pretty) {
			tokenConsumer = new JsPrettyPrinter(concatenator);
		} else {
			tokenConsumer = new JsMinimalPrinter(concatenator);
		}

		renderContext = new RenderContext(tokenConsumer).withEmbeddable(true).withJson(false);
		cajoledModule.render(renderContext);
		renderContext.getOut().noMoreTokens();

		if (renderCallback.exception != null) {
			throw renderCallback.exception;
		}

		source = stringBuilder.toString();
		resp.setContentType("text/plain; charset=utf-8");
		resp.getWriter().println(source);

	}

}