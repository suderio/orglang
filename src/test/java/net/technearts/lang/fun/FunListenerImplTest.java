package net.technearts.lang.fun;

import org.antlr.v4.runtime.CharStream;
import org.antlr.v4.runtime.CharStreams;
import org.antlr.v4.runtime.CommonTokenStream;
import org.antlr.v4.runtime.ConsoleErrorListener;
import org.antlr.v4.runtime.atn.PredictionMode;
import org.antlr.v4.runtime.tree.ParseTree;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;

import java.math.BigInteger;

import static net.technearts.lang.fun.TestUtils.assertAll;
import static net.technearts.lang.fun.TestUtils.assertNumbersEqual;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class FunListenerImplTest {

  private Table evaluate(String code) {
    CharStream input = CharStreams.fromString(code);
    FunLexer lexer = new FunLexer(input);
    lexer.addErrorListener(ConsoleErrorListener.INSTANCE);
    CommonTokenStream tokens = new CommonTokenStream(lexer);
    FunParser parser = new FunParser(tokens);
    parser.addErrorListener(ConsoleErrorListener.INSTANCE);
    parser.getInterpreter().setPredictionMode(PredictionMode.LL_EXACT_AMBIG_DETECTION);
    ExecutionEnvironment env = new ExecutionEnvironment();
    if (env.isDebug())
      parser.addParseListener(new FunListenerImpl());
    ParseTree tree = parser.file();
    FunVisitorImpl visitor = new FunVisitorImpl(env);
    System.out.println(tree.toStringTree(parser));
    var result = (Table) visitor.visit(tree);
    System.out.println(result);
    return result;
  }

  @Test
  void testAssignments() {
    var table = evaluate("""
                            x : 42;
                            y : x + 8;
                            x = 42;
                            y = 50;
        """);
    assertAll(table);
  }

  @Test
  void testUnaryOperators() {
    var table = evaluate("""
                            x : -10;
                            y : +20;
                            z : ~false;
                            -11 = --x;
                            4 = ++ [1 2 y];
                            ~~8;
        """);

    assertAll(table);
  }

  @Test
  void testArithmeticExpressions() {
    var table = evaluate("""
        14 = 2 + 3 * 4;
        20 = (2 + 3) * 4;
        100 = 10 ** 2;
        """);

    assertAll(table);
  }

  @Test
  void testLogicalExpressions() {
    var table = evaluate("""
                            ~(true && false);
                            true || false;
                            true ^ false;
        """);

    assertAll(table);
  }

  @Test
  void testComparisons() {
    var table = evaluate("""
                            10 = 10;
                            10 <> 5;
                            10 > [5];
                            10 >= 10;
                            10 < 15;
                            10 <= 10;
        """);

    assertAll(table);
  }

  @Test
  void testNullTestExpressions() {
    var table = evaluate("""
        42 = (null ?? 42);
        10 = (10 ?? 42);
        """);

    assertAll(table);
  }

  @Test
  void testTestExpressions() {
    var table = evaluate("""
                            null ? ?? true;
                            10 ?;
                            ~(1 = 0) ?;
        """);

    assertAll(table);
  }

  @Test
  void testTableConstruction() {
    var table = (Table) evaluate("""
        t : [1 2 "hello" 3];
        """).get(BigInteger.ZERO);
    assertNumbersEqual(1, table.get(BigInteger.ZERO));
    assertNumbersEqual(2, table.get(BigInteger.ONE));
    assertEquals("hello", table.get(BigInteger.TWO));
  }

  @Test
  void testTableConcatenation() {
    var table = (Table) evaluate("""
        t : 1, 2, 3, 4;
        """).get(BigInteger.ZERO);
    assertNumbersEqual(1, table.get(BigInteger.ZERO));
    assertNumbersEqual(2, table.get(BigInteger.ONE));
    assertNumbersEqual(3, table.get(BigInteger.TWO));
    assertNumbersEqual(4, table.get(BigInteger.valueOf(3)));
  }

  @Test
  void testDereferencing() {
    var table = evaluate("""
                            t : [1 2 3 "4"];
                            2 = t.1;
                            "4" = t."1234".3;
        """);

    assertAll(table);
  }

  @Test
  void testNestedExpressions() {
    var table = evaluate("""
                            70 = 10 + (20 * (5 - 2));
        """);

    assertAll(table);
  }

  @Test
  void testCallOperators() {
    var table = evaluate("""
                            negate : { -right };
                            pos : { +right };
                            inv : { ~right };
                            -10 = negate 10;
                            20 = pos 20;
                            false = inv true;
        """);
    assertAll(table);
  }

  @Test
  void testShift() {
    var table = evaluate("""
                            64 = 16 << 2;
                            16 = 128 >> 3;
        """);
    assertAll(table);
  }

  @Test
  void testOperators() {
    var table = evaluate("""
                            sq: { right * right };
                            16 = sq 4;
        """);
    assertAll(table);
  }

  @Test
  void testRange() {
    var table = evaluate("""
        5 = +(1..5);
        4 = +(8..5);
        1 = +(3..3);
        """);
    assertAll(table);
  }

  @Test
  void testAssignOpExpression() {
    var table = evaluate("""
            x : 10;
            y : 20;
            x :+ 5;
            y :* 2;
            x = 15;
            y = 40;
        """);
    assertAll(table);
  }

  @Test
  void testSubstExpression() {
    var table = evaluate("""
                            x : 10;
                            y : "x: ${x}";
                            z : "0: $0, 1: $1" $ [3 6];
                            "x: 10" = y;
                            "0: 3, 1: 6" $ [3 6] = z;
        """);
    assertAll(table);
  }

  @Test
  void testFibonacci() {
    var table = evaluate("""
                            fib : { [1 1].right ?? (this(right - 1)) + (this(right - 2)) };
                            1 = fib 0;
                            8 = fib 5;
        """);

    assertAll(table);
  }

  @Test
  void testFatorial() {
    var table = evaluate("""
                            fat : { [1 1].right ?? (this(right - 1)) * right };
                            1 = (fat 1);
                            6 = (fat 3);
                            120 = (fat 5);
        """);

    assertAll(table);
  }

  @Test
  void testTableDerefTable() {
    var table = evaluate("""
                            x: [1 2 3 4 5];
                            y: [1 2 3];
                            1 = x.y.0;
                            2 = x.y.1;
                            3 = x.y.2;
        """);
    assertAll(table);
  }

  @Test
  void testTableDerefOp() {
    var table = evaluate("""
                            x: [1 2 3 "4"];
                            y: {right < 2 && right > 0};
                            2 = x.y.0;
                            1 = y.x.0;
                            "4" = y.x.1;
                            "4" = [1 2 3 "4"].{right > 2}.0;
                            3 = {right > 2}.[1 2 3 "4"].0;
        """);
    assertAll(table);
  }

  @Test
  void testBiOp() {
    var table = evaluate("""
                            idiv: {right / left};
                            x: (3 idiv 0);
                            0 = x;
        """);
    assertAll(table);
  }

  @Test
  void testWalker() {
    Main main = new Main();
    System.out.printf(" # NPR: %s\n", main.walk("(-1 + 2 * (3 - 4) / 5 > (0,1,2));"));
    System.out.printf(" # NPR: %s\n", main.walk("a : 3;"));
    System.out.printf(" # NPR: %s\n", main.walk("{ (1, 1).right ?? (this(right - 1)) * right };"));
  }

}
