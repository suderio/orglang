package net.technearts.lang.fun;

import org.antlr.v4.runtime.tree.ParseTree;
import org.antlr.v4.runtime.tree.ParseTreeProperty;

public class FunListenerImpl extends FunBaseListener {

    private final ParseTreeProperty<String> values = new ParseTreeProperty<>();


    @Override
    public void exitFileTable(FunParser.FileTableContext ctx) {
        super.exitFileTable(ctx);
    }

    @Override
    public void exitExpressionExp(FunParser.ExpressionExpContext ctx) {
        super.exitExpressionExp(ctx);
    }

    @Override
    public void exitOperatorExp(FunParser.OperatorExpContext ctx) {
        setValue(ctx, getValue(ctx.expression()));
    }

    @Override
    public void exitShiftExp(FunParser.ShiftExpContext ctx) {
        infix(ctx);
    }

    private void infix(FunParser.ExpressionContext ctx) {
        String left = getValue(ctx.getRuleContext(FunParser.ExpressionContext.class,0));
        String right = getValue(ctx.getRuleContext(FunParser.ExpressionContext.class,1));
        String operator = ctx.getChild(1).getText(); // "+" ou "-"
        setValue(ctx, left + " " + right + " " + operator);
    }

    private void prefix(FunParser.ExpressionContext ctx) {
        String right = getValue(ctx.getRuleContext(FunParser.ExpressionContext.class,0));
        String operator = ctx.getChild(0).getText(); // "+" ou "-"
        setValue(ctx, right + " " + operator);
    }

    private void postfix(FunParser.ExpressionContext ctx) {
        String left = getValue(ctx.getRuleContext(FunParser.ExpressionContext.class,0));
        String operator = ctx.getChild(1).getText(); // "+" ou "-"
        setValue(ctx, left + " " + operator);
    }

    // Método auxiliar para recuperar valores
    private String getValue(ParseTree ctx) {
        return values.get(ctx);
    }

    // Método auxiliar para definir valores pós-fixos para os nós da árvore
    private void setValue(ParseTree ctx, String value) {
        values.put(ctx, value);
    }

    @Override
    public void exitAssignExp(FunParser.AssignExpContext ctx) {
        String left = getValue(ctx.ID());
        String right = getValue(ctx.right);
        String operator = ctx.ASSIGN().getText();
        setValue(ctx, left + " " + right + " " + operator);
    }

    @Override
    public void exitAndExp(FunParser.AndExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitOrShortExp(FunParser.OrShortExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitNullTestExp(FunParser.NullTestExpContext ctx) {
        String left = getValue(ctx.getRuleContext(FunParser.ExpressionContext.class,0));
        String right = getValue(ctx.getRuleContext(FunParser.ExpressionContext.class,1));
        String operator = ctx.getChild(1).getText(); // "+" ou "-"
        setValue(ctx, left + " " + operator + " " + right);
    }

    @Override
    public void exitXorExp(FunParser.XorExpContext ctx) {
        infix(ctx);
    }
    @Override
    public void exitThisExp(FunParser.ThisExpContext ctx) {
        prefix(ctx);
    }

    @Override
    public void exitDerefExp(FunParser.DerefExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitSubstExp(FunParser.SubstExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitAndShortExp(FunParser.AndShortExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitTestExp(FunParser.TestExpContext ctx) {
        postfix(ctx);
    }

    @Override
    public void exitKeyValueExp(FunParser.KeyValueExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitBiCallExp(FunParser.BiCallExpContext ctx) {
        infix(ctx);
    }

    // TODO
    @Override
    public void exitTableConstructExp(FunParser.TableConstructExpContext ctx) {
        super.exitTableConstructExp(ctx);
    }

    @Override
    public void exitEqualityExp(FunParser.EqualityExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitOrExp(FunParser.OrExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitCallExp(FunParser.CallExpContext ctx) {
        prefix(ctx);
    }

    @Override
    public void exitAssignOpExp(FunParser.AssignOpExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitRangeExp(FunParser.RangeExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitUnaryExp(FunParser.UnaryExpContext ctx) {
        prefix(ctx);
    }

    @Override
    public void exitComparisonExp(FunParser.ComparisonExpContext ctx) {
        infix(ctx);
    }

    @Override
    public void exitTableConcatSepExp(FunParser.TableConcatSepExpContext ctx) {
        infix(ctx);
    }

    // Notação pós-fixa para expressões entre parênteses
    @Override
    public void exitParenthesisExp(FunParser.ParenthesisExpContext ctx) {
        setValue(ctx, getValue(ctx.expression()));
    }

    // Notação pós-fixa para expressões de adição/subtração
    @Override
    public void exitAddSubExp(FunParser.AddSubExpContext ctx) {
        infix(ctx);
    }


    // Notação pós-fixa para expressões de multiplicação/divisão/módulo
    @Override
    public void exitMulDivModExp(FunParser.MulDivModExpContext ctx) {
        infix(ctx);
    }

    // Notação pós-fixa para expressões de exponenciação
    @Override
    public void exitPowerExp(FunParser.PowerExpContext ctx) {
        infix(ctx);
    }

    // Notação pós-fixa para expressões de números inteiros
    @Override
    public void exitIntegerLiteral(FunParser.IntegerLiteralContext ctx) {
        setValue(ctx, ctx.INTEGER().getText());
    }

    // Notação pós-fixa para identificadores
    @Override
    public void exitIdAtomExp(FunParser.IdAtomExpContext ctx) {
        setValue(ctx, ctx.ID().getText());
    }

    @Override
    public void exitDecimalLiteral(FunParser.DecimalLiteralContext ctx) {
        setValue(ctx, ctx.DECIMAL().getText());
    }

    @Override
    public void exitRightAtomLiteral(FunParser.RightAtomLiteralContext ctx) {
        setValue(ctx, ctx.RIGHT().getText());
    }


    @Override
    public void exitTrueLiteral(FunParser.TrueLiteralContext ctx) {
        setValue(ctx, ctx.TRUE().getText());
    }

    @Override
    public void exitFalseLiteral(FunParser.FalseLiteralContext ctx) {
        setValue(ctx, ctx.FALSE().getText());
    }

    @Override
    public void exitLeftAtomLiteral(FunParser.LeftAtomLiteralContext ctx) {
        setValue(ctx, ctx.LEFT().getText());
    }

    @Override
    public void exitNullLiteral(FunParser.NullLiteralContext ctx) {
        setValue(ctx, ctx.NULL().getText());
    }

    @Override
    public void exitDocStringLiteral(FunParser.DocStringLiteralContext ctx) {
        setValue(ctx, ctx.DOCSTRING().getText());
    }

    @Override
    public void exitStringLiteral(FunParser.StringLiteralContext ctx) {
        setValue(ctx, ctx.SIMPLESTRING().getText());
    }

    // Método para obter a saída final pós-fixa da expressão
    public String getPostfix(ParseTree tree) {
        return getValue(tree);
    }
}
