// Public Domain (-) 2010-2011 The Ampify Authors.
// See the UNLICENSE file for details.

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
		resp.getWriter().print("ERROR: " + errorMessage);
	}

	public void doGet(HttpServletRequest req, HttpServletResponse resp)
		throws IOException {
		resp.sendRedirect("/");
	}

	public void doPost(HttpServletRequest req, HttpServletResponse resp)
		throws IOException {

		this.resp = resp;

		String prettyParam, source, inputType;

		prettyParam = req.getParameter("pretty");
		source = req.getParameter("source");
		inputType = req.getParameter("input_type");

		resp.setContentType("text/plain; charset=utf-8");

		if (source == null) {
			badRequest("The `source` parameter was not specified.");
			return;
		}

		if (source.length() > 102400) { // 100kB
			badRequest("The `source` parameter value is too long!");
			return;
		}

		if (inputType == null) {
			badRequest("The `input_type` parameter was not specified.");
			return;
		}

		if (inputType.equals("coffee")) {
			try {
				source = getCompiler().compile(source);
			} catch (JavaScriptException e) {
				error(e.getValue().toString());
				return;
			}
			if (req.getParameter("coffee") != null) {
				resp.getWriter().print(source);
				return;
			}
		} else {
			if (!inputType.equals("js")) {
				badRequest("Unknown `input_type` parameter value.");
				return;
			}
		}

		BuildInfo buildInfo;
		CajoledModule cajoledModule = null;
		CharProducer inputStream;
		MessageQueue messageQueue = new SimpleMessageQueue();
		Block parsedBlock;
		Rewriter rewriter;
		JsTokenQueue tokenQueue;
		UncajoledModule uncajoledModule;

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

			StringBuilder errorMessage;
			String inputSourceString;

			cajoledModule = null;
			errorMessage = new StringBuilder("Couldn't cajole the source.\n\n");
			inputSourceString = InputSource.PREDEFINED.toString() + ":";
			for (Message m: messageQueue.getMessages()) {
				errorMessage.append(m.toString().replace(inputSourceString, "Line: "));
				errorMessage.append("\n\n");
			}
			error(errorMessage.toString());
			return;

		}

		boolean pretty;
		Concatenator concatenator;
		IOCallback renderCallback;
		RenderContext renderContext;
		StringBuilder stringBuilder;
		TokenConsumer tokenConsumer;

		pretty = false;
		if ((prettyParam != null) && prettyParam.equals("1")) {
			pretty = true;
		}

		renderCallback = new IOCallback();
		stringBuilder = new StringBuilder();
		concatenator = new Concatenator(stringBuilder, renderCallback);

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
		resp.getWriter().print(source);

	}

}