import numpy as np
from onnxruntime import InferenceSession
from skl2onnx import convert_sklearn
from skl2onnx.common.data_types import DoubleTensorType


def convert_to_onnx(model):
    initial_types = [('double_input', DoubleTensorType([None, 13]))]
    onx = convert_sklearn(model, initial_types=initial_types, target_opset=8)
    return onx


# deprecated
def match_prediction_with_id(ids, pred):
    r = dict()
    for i, p in enumerate(pred):
        r[ids.iloc[i]] = p['True']
    return r


class Model:
    def __init__(self, model):
        self.sess = InferenceSession(model)
        self.input_name = self.sess.get_inputs()[0].name
        self.output_name = self.sess.get_outputs()[1].name

    def predict(self, X):
        return self.sess.run([self.output_name], {self.input_name: X.to_numpy().astype(np.double)})[0]
